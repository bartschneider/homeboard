#!/usr/bin/env python3
"""
Todo list widget for E-Paper Dashboard
Displays a simple todo list from a text file or JSON source
"""

import json
import sys
import os
from typing import Dict, Any, List


def load_parameters() -> Dict[str, Any]:
    """Load and parse widget parameters from command line argument"""
    if len(sys.argv) < 2:
        return {}
    
    try:
        return json.loads(sys.argv[1])
    except (json.JSONDecodeError, IndexError):
        return {}


def load_todos_from_file(file_path: str) -> List[Dict[str, Any]]:
    """Load todos from a file (supports .txt and .json formats)"""
    if not os.path.exists(file_path):
        return [{"error": f"Todo file not found: {file_path}"}]
    
    try:
        with open(file_path, 'r', encoding='utf-8') as f:
            content = f.read().strip()
        
        if file_path.endswith('.json'):
            # JSON format: [{"task": "...", "done": false, "priority": "high"}, ...]
            todos = json.loads(content)
            if not isinstance(todos, list):
                return [{"error": "JSON file must contain a list of todos"}]
            return todos
        
        elif file_path.endswith('.txt'):
            # Text format: each line is a todo, prefix with [x] for done
            todos = []
            for line_num, line in enumerate(content.split('\n'), 1):
                line = line.strip()
                if not line:
                    continue
                
                done = False
                if line.startswith('[x]') or line.startswith('[X]'):
                    done = True
                    task = line[3:].strip()
                elif line.startswith('[ ]'):
                    done = False
                    task = line[3:].strip()
                else:
                    task = line
                
                todos.append({
                    "task": task,
                    "done": done,
                    "priority": "normal",
                    "line": line_num
                })
            
            return todos
        
        else:
            return [{"error": f"Unsupported file format: {file_path}"}]
    
    except json.JSONDecodeError as e:
        return [{"error": f"Invalid JSON format: {str(e)}"}]
    except Exception as e:
        return [{"error": f"Error reading file: {str(e)}"}]


def get_static_todos() -> List[Dict[str, Any]]:
    """Return a static list of example todos"""
    return [
        {"task": "Check server monitoring", "done": False, "priority": "high"},
        {"task": "Update dashboard widgets", "done": False, "priority": "normal"},
        {"task": "Backup configuration files", "done": True, "priority": "normal"},
        {"task": "Review system logs", "done": False, "priority": "low"},
        {"task": "Plan next homelab upgrade", "done": False, "priority": "low"},
    ]


def generate_html(todos: List[Dict[str, Any]], max_items: int) -> str:
    """Generate HTML output for the todo widget"""
    if len(todos) == 1 and "error" in todos[0]:
        return f"""
        <div class="todo-widget">
            <h2>📝 Todo List</h2>
            <div style="text-align: center; color: #666;">
                ⚠️ Error: {todos[0]['error']}<br>
                <small>Check file path and format</small>
            </div>
        </div>
        """
    
    if not todos:
        return f"""
        <div class="todo-widget">
            <h2>📝 Todo List</h2>
            <div style="text-align: center; color: #666;">
                No todos found
            </div>
        </div>
        """
    
    # Sort todos: incomplete first, then by priority
    priority_order = {"high": 1, "normal": 2, "low": 3}
    sorted_todos = sorted(
        todos, 
        key=lambda x: (x.get("done", False), priority_order.get(x.get("priority", "normal"), 2))
    )
    
    # Limit number of displayed items
    display_todos = sorted_todos[:max_items]
    
    html_parts = []
    html_parts.append('<div class="todo-widget">')
    html_parts.append('<h2>📝 Todo List</h2>')
    
    # Count statistics
    total_todos = len(todos)
    completed_todos = sum(1 for todo in todos if todo.get("done", False))
    pending_todos = total_todos - completed_todos
    
    # Add summary
    html_parts.append(f'<div style="font-size: 0.9em; margin-bottom: 10px; text-align: center;">')
    html_parts.append(f'{pending_todos} pending • {completed_todos} completed • {total_todos} total')
    html_parts.append('</div>')
    
    # Add todo items
    for todo in display_todos:
        task = todo.get("task", "Unknown task")
        done = todo.get("done", False)
        priority = todo.get("priority", "normal")
        
        # Determine checkbox and styling
        checkbox = "☑️" if done else "☐"
        task_style = "text-decoration: line-through; color: #888;" if done else ""
        
        # Priority indicator
        priority_icon = ""
        if priority == "high" and not done:
            priority_icon = "🔴 "
        elif priority == "low":
            priority_icon = "🔵 "
        
        html_parts.append(f'''
        <div style="margin: 5px 0; display: flex; align-items: flex-start; {task_style}">
            <span style="margin-right: 5px; flex-shrink: 0;">{checkbox}</span>
            <span style="flex: 1; word-wrap: break-word;">{priority_icon}{task}</span>
        </div>
        ''')
    
    # Show truncation indicator if needed
    if len(todos) > max_items:
        html_parts.append(f'<div style="text-align: center; font-size: 0.8em; color: #666; margin-top: 8px;">')
        html_parts.append(f'... and {len(todos) - max_items} more items')
        html_parts.append('</div>')
    
    html_parts.append('</div>')
    
    return ''.join(html_parts)


def main():
    """Main widget execution function"""
    try:
        # Load parameters
        params = load_parameters()
        
        # Get configuration with defaults
        file_path = params.get("file_path", "")
        max_items = params.get("max_items", 8)
        use_static = params.get("use_static", False)
        
        # Load todos
        if use_static or not file_path:
            todos = get_static_todos()
        else:
            todos = load_todos_from_file(file_path)
        
        # Generate and output HTML
        html_output = generate_html(todos, max_items)
        print(html_output)
        
    except Exception as e:
        # Error handling - output user-friendly error message
        error_html = f"""
        <div class="todo-widget">
            <h2>📝 Todo List</h2>
            <div style="text-align: center; color: #666;">
                ⚠️ Error: {str(e)}<br>
                <small>Check widget configuration</small>
            </div>
        </div>
        """
        print(error_html)


if __name__ == "__main__":
    main()