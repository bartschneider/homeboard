#!/usr/bin/env python3
"""
Widget Builder Agent using Google Agent Development Kit (ADK)
This implements the proper ADK patterns for interactive widget building.
"""

import os
import asyncio
from typing import Dict, List, Optional, Any
from dataclasses import dataclass
import json
import logging

# Google ADK imports
from google.adk.agents import Agent, LlmAgent, SequentialAgent
from google.adk.runners import Runner
from google.adk.sessions import InMemorySessionService, Session
from google.adk.events import Event
from google.adk.agents.invocation_context import InvocationContext

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

@dataclass
class WidgetConfig:
    """Widget configuration data structure"""
    name: str = ""
    api_url: str = ""
    template_type: str = ""
    description: str = ""
    data_mapping: Dict[str, Any] = None
    
    def __post_init__(self):
        if self.data_mapping is None:
            self.data_mapping = {}

class WidgetBuilderAgent:
    """Main Widget Builder Agent using Google ADK"""
    
    def __init__(self, gemini_api_key: str):
        self.gemini_api_key = gemini_api_key
        self.session_service = InMemorySessionService()
        self.app_name = "widget_builder"
        
        # Initialize ADK agents
        self._setup_agents()
        self._setup_runner()
        
    def _setup_agents(self):
        """Setup specialized ADK agents for widget building"""
        
        # API Analyzer Agent
        self.api_analyzer = LlmAgent(
            name="APIAnalyzer",
            model="gemini-2.0-flash",
            description="Analyzes API endpoints and extracts schema information",
            instruction="""
            You are an API analysis specialist. When given an API URL:
            1. Analyze the endpoint structure and authentication requirements
            2. Identify the data schema and available fields
            3. Suggest optimal data mappings for widget templates
            4. Validate API accessibility and response format
            
            Respond with structured analysis including:
            - API validation status
            - Available data fields
            - Suggested template types
            - Recommended field mappings
            """,
            output_key="api_analysis"
        )
        
        # Widget Designer Agent  
        self.widget_designer = LlmAgent(
            name="WidgetDesigner", 
            model="gemini-2.0-flash",
            description="Designs optimal widget configurations based on requirements",
            instruction="""
            You are a widget design specialist. Based on user requirements and API analysis:
            1. Select the most appropriate widget template
            2. Configure optimal field mappings
            3. Set appropriate refresh intervals and caching
            4. Ensure responsive design for e-paper displays
            
            Consider:
            - Data update frequency
            - Display constraints (e-paper optimization)
            - User experience and readability
            - Performance implications
            """,
            output_key="widget_design"
        )
        
        # Natural Language Processor
        self.nl_processor = LlmAgent(
            name="NLProcessor",
            model="gemini-2.0-flash", 
            description="Converts natural language descriptions to structured requirements",
            instruction="""
            You are a natural language processing specialist for widget requirements.
            Convert user descriptions into structured widget specifications:
            
            Extract:
            - Widget type and purpose
            - Data source requirements
            - Display preferences
            - Update frequency needs
            - Special requirements (authentication, filtering, etc.)
            
            Output structured requirements that other agents can use.
            """,
            output_key="nl_requirements"
        )
        
        # Validation Agent
        self.validator = LlmAgent(
            name="ValidationAgent",
            model="gemini-2.0-flash",
            description="Validates widget configurations and performs quality checks",
            instruction="""
            You are a widget validation specialist. Review widget configurations for:
            1. Technical correctness and feasibility
            2. Security considerations (API keys, data exposure)
            3. Performance optimization
            4. User experience quality
            5. E-paper display compatibility
            
            Provide validation results with specific recommendations for improvements.
            """,
            output_key="validation_results"
        )
        
        # Coordinator Agent - orchestrates the workflow
        self.coordinator = LlmAgent(
            name="WidgetBuilderCoordinator",
            model="gemini-2.0-flash",
            description="Main coordinator for widget building workflow",
            instruction="""
            You are the main widget builder coordinator. Guide users through the widget creation process:
            
            1. Understand user requirements through natural conversation
            2. Coordinate with specialist agents for analysis and design
            3. Provide clear, helpful responses and next steps
            4. Manage the workflow from discovery to completion
            
            Always:
            - Ask clarifying questions when needed
            - Explain what you're doing and why
            - Provide actionable next steps
            - Offer specific suggestions and examples
            
            Keep responses conversational and helpful.
            """,
            sub_agents=[self.api_analyzer, self.widget_designer, self.nl_processor, self.validator],
            output_key="coordinator_response"
        )
        
        logger.info("✅ ADK agents initialized successfully")
        
    def _setup_runner(self):
        """Setup ADK runner with session management"""
        self.runner = Runner(
            agent=self.coordinator,
            app_name=self.app_name,
            session_service=self.session_service
        )
        logger.info("✅ ADK runner initialized successfully")
        
    async def process_message(self, session_id: str, user_id: str, message: str, context: Dict[str, Any] = None) -> Dict[str, Any]:
        """Process user message using ADK conversation flow"""
        try:
            # Get or create session
            session = await self.session_service.get_session(
                app_name=self.app_name,
                user_id=user_id, 
                session_id=session_id
            )
            
            if session is None:
                session = await self.session_service.create_session(
                    app_name=self.app_name,
                    user_id=user_id,
                    session_id=session_id
                )
                logger.info(f"Created new ADK session: {session_id}")
            
            # Add context to session state if provided
            if context:
                for key, value in context.items():
                    session.state[key] = value
            
            # Process message through ADK
            response = await self.runner.run_async(
                user_input=message,
                session_id=session_id,
                user_id=user_id
            )
            
            # Extract response data
            events = response.events if hasattr(response, 'events') else []
            final_response = ""
            actions = []
            
            # Process events to extract agent responses and actions
            for event in events:
                if hasattr(event, 'content') and event.content:
                    final_response += event.content + "\n"
                    
                if hasattr(event, 'actions') and event.actions:
                    actions.extend(event.actions)
            
            # Get updated session state
            updated_session = await self.session_service.get_session(
                app_name=self.app_name,
                user_id=user_id,
                session_id=session_id
            )
            
            # Extract widget configuration from session state
            widget_config = self._extract_widget_config(updated_session.state)
            
            # Determine current phase
            phase = self._determine_phase(updated_session.state, message)
            
            return {
                "response": final_response.strip(),
                "actions": actions,
                "session_state": dict(updated_session.state),
                "widget_config": widget_config.__dict__ if widget_config else None,
                "phase": phase,
                "suggestions": self._generate_suggestions(phase, updated_session.state)
            }
            
        except Exception as e:
            logger.error(f"Error processing message: {e}")
            return {
                "response": f"I encountered an error: {str(e)}. Let me try to help you differently.",
                "actions": [],
                "session_state": {},
                "widget_config": None,
                "phase": "discovery",
                "suggestions": ["Try describing your widget requirements again"]
            }
    
    def _extract_widget_config(self, state: Dict[str, Any]) -> Optional[WidgetConfig]:
        """Extract widget configuration from session state"""
        try:
            config = WidgetConfig()
            
            # Extract from various possible state keys
            if "widget_name" in state:
                config.name = state["widget_name"]
            if "api_url" in state:
                config.api_url = state["api_url"]  
            if "template_type" in state:
                config.template_type = state["template_type"]
            if "description" in state:
                config.description = state["description"]
            if "data_mapping" in state:
                config.data_mapping = state["data_mapping"]
                
            # Check if we have enough info for a valid config
            if config.name or config.api_url or config.template_type:
                return config
                
        except Exception as e:
            logger.error(f"Error extracting widget config: {e}")
            
        return None
    
    def _determine_phase(self, state: Dict[str, Any], message: str) -> str:
        """Determine current workflow phase"""
        message_lower = message.lower()
        
        # Check for completion indicators
        if "save" in message_lower or "create" in message_lower or "done" in message_lower:
            return "completion"
            
        # Check for validation phase
        if "test" in message_lower or "validate" in message_lower or "preview" in message_lower:
            return "validation"
            
        # Check for configuration phase  
        if state.get("template_type") or state.get("api_analysis"):
            return "configuration"
            
        # Default to discovery
        return "discovery"
    
    def _generate_suggestions(self, phase: str, state: Dict[str, Any]) -> List[str]:
        """Generate contextual suggestions based on phase"""
        if phase == "discovery":
            return [
                "Tell me what kind of data you want to display",
                "Provide an API URL for automatic analysis", 
                "Describe your dashboard requirements"
            ]
        elif phase == "configuration":
            return [
                "Adjust the widget settings",
                "Test the API connection",
                "Preview the widget layout"
            ]
        elif phase == "validation":
            return [
                "Save the widget configuration",
                "Make final adjustments",
                "Test with live data"
            ]
        elif phase == "completion":
            return [
                "Create another widget",
                "Go to dashboard builder",
                "View all widgets"
            ]
        else:
            return ["How can I help you with your widget?"]

# Flask API wrapper for the ADK agent
from flask import Flask, request, jsonify
from flask_cors import CORS

app = Flask(__name__)
CORS(app)

# Initialize ADK agent
GEMINI_API_KEY = os.getenv("GEMINI_API_KEY")
if not GEMINI_API_KEY:
    raise ValueError("GEMINI_API_KEY environment variable is required")

widget_agent = WidgetBuilderAgent(GEMINI_API_KEY)

@app.route('/chat', methods=['POST'])
async def chat():
    """Chat endpoint using ADK"""
    try:
        data = request.get_json()
        
        session_id = data.get('session_id', 'default')
        user_id = data.get('user_id', 'default_user')
        message = data.get('message', '')
        context = data.get('context', {})
        
        if not message:
            return jsonify({"error": "Message is required"}), 400
            
        response = await widget_agent.process_message(session_id, user_id, message, context)
        
        return jsonify(response)
        
    except Exception as e:
        logger.error(f"Chat endpoint error: {e}")
        return jsonify({"error": str(e)}), 500

@app.route('/health', methods=['GET'])
def health():
    """Health check endpoint"""
    return jsonify({"status": "healthy", "service": "ADK Widget Builder"})

if __name__ == '__main__':
    # Run the Flask app
    app.run(host='0.0.0.0', port=5000, debug=True)