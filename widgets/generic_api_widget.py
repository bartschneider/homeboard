#!/usr/bin/env python3
"""
Generic API Widget Executor

This script executes widgets based on JSON configuration instead of hardcoded Python scripts.
It fetches data from APIs, applies data mapping, and renders HTML using predefined templates.

Usage:
    python3 generic_api_widget.py '{"api_url": "...", "template_type": "...", "data_mapping": {...}}'
"""

import json
import sys
import requests
import time
from typing import Any, Dict, List, Optional, Union
from datetime import datetime
import html


class WidgetExecutor:
    """Generic widget executor that processes API data using configuration."""
    
    def __init__(self, config: Dict[str, Any]):
        self.config = config
        self.data_source = config.get('data_source', 'api')
        self.api_url = config.get('api_url', '')
        self.api_headers = config.get('api_headers', {})
        self.template_type = config.get('template_type', 'key_value')
        self.data_mapping = config.get('data_mapping', {})
        self.rss_config = config.get('rss_config', {})
        self.timeout = config.get('timeout', 30)
        
    def execute(self) -> str:
        """Execute the widget and return HTML output."""
        try:
            # Fetch data based on data source type
            if self.data_source == 'rss':
                data = self.fetch_rss_data()
            else:
                data = self.fetch_api_data()
            
            # Apply data mapping
            mapped_data = self.apply_data_mapping(data)
            
            # Render HTML using template
            html_output = self.render_template(mapped_data)
            
            return html_output
            
        except Exception as e:
            return self.render_error(f"Widget execution failed: {str(e)}")
    
    def fetch_api_data(self) -> Dict[str, Any]:
        """Fetch data from the configured API endpoint."""
        if not self.api_url:
            raise ValueError("API URL is required")
        
        headers = dict(self.api_headers) if self.api_headers else {}
        headers.setdefault('User-Agent', 'E-Paper-Dashboard/1.0')
        
        try:
            response = requests.get(
                self.api_url,
                headers=headers,
                timeout=self.timeout
            )
            response.raise_for_status()
            return response.json()
            
        except requests.exceptions.RequestException as e:
            raise Exception(f"API request failed: {str(e)}")
        except json.JSONDecodeError as e:
            raise Exception(f"Invalid JSON response: {str(e)}")
    
    def fetch_rss_data(self) -> Dict[str, Any]:
        """Fetch data from RSS feed by calling the RSS API endpoint."""
        if not self.api_url:
            raise ValueError("RSS feed URL is required")
        
        # Use the RSS preview API endpoint to fetch parsed RSS data
        try:
            rss_api_url = "http://localhost:8080/api/rss/preview"
            payload = {
                "feed_url": self.api_url,
                "rss_config": self.rss_config
            }
            
            response = requests.post(
                rss_api_url,
                json=payload,
                headers={"Content-Type": "application/json"},
                timeout=self.timeout
            )
            response.raise_for_status()
            
            data = response.json()
            if 'feed' not in data:
                raise Exception("Invalid RSS API response format")
            
            return data['feed']
            
        except requests.exceptions.ConnectionError:
            # Fallback to direct RSS parsing if API is not available
            return self.fetch_rss_direct()
        except Exception as e:
            # Handle any other exception that might occur during API call
            if "Connection" in str(e) or "connection" in str(e).lower():
                return self.fetch_rss_direct()
            raise Exception(f"RSS API request failed: {str(e)}")
        except json.JSONDecodeError as e:
            raise Exception(f"Invalid RSS API response: {str(e)}")
    
    def fetch_rss_direct(self) -> Dict[str, Any]:
        """Direct RSS parsing fallback when API is not available."""
        try:
            import xml.etree.ElementTree as ET
            from urllib.parse import urlparse
            
            response = requests.get(
                self.api_url,
                headers={"User-Agent": "E-Paper-Dashboard/1.0 RSS Reader"},
                timeout=self.timeout
            )
            response.raise_for_status()
            
            # Parse RSS XML
            root = ET.fromstring(response.content)
            
            # Extract basic RSS data
            channel = root.find('channel')
            if channel is None:
                raise Exception("Invalid RSS format: no channel element found")
            
            feed_data = {
                "title": self.get_element_text(channel, 'title'),
                "description": self.get_element_text(channel, 'description'),
                "link": self.get_element_text(channel, 'link'),
                "items": []
            }
            
            # Extract items
            max_items = self.rss_config.get('max_items', 10)
            for item in channel.findall('item')[:max_items]:
                item_data = {
                    "title": self.get_element_text(item, 'title'),
                    "description": self.get_element_text(item, 'description'),
                    "link": self.get_element_text(item, 'link'),
                    "pub_date": self.get_element_text(item, 'pubDate'),
                    "author": self.get_element_text(item, 'author'),
                    "guid": self.get_element_text(item, 'guid'),
                }
                feed_data["items"].append(item_data)
            
            return feed_data
            
        except ET.ParseError as e:
            raise Exception(f"RSS XML parsing failed: {str(e)}")
        except requests.exceptions.RequestException as e:
            raise Exception(f"RSS request failed: {str(e)}")
    
    def get_element_text(self, parent, tag_name: str) -> str:
        """Helper to safely extract text from XML element."""
        element = parent.find(tag_name)
        return element.text.strip() if element is not None and element.text else ""
    
    def apply_data_mapping(self, api_data: Dict[str, Any]) -> Dict[str, Any]:
        """Apply data mapping configuration to extract values from API response."""
        mapped_data = {}
        
        for field_name, json_path in self.data_mapping.items():
            try:
                value = self.extract_json_path(api_data, json_path)
                mapped_data[field_name] = value
            except Exception as e:
                print(f"Warning: Failed to extract '{json_path}' for field '{field_name}': {e}", file=sys.stderr)
                mapped_data[field_name] = None
        
        return mapped_data
    
    def extract_json_path(self, data: Any, path: str) -> Any:
        """Extract value from data using JSONPath-like syntax."""
        if not path:
            return data
        
        # If path is a simple literal value (like "5"), try to parse it as such
        if '.' not in path and '[' not in path and ']' not in path:
            # Check if it's a simple key access
            if isinstance(data, dict) and path in data:
                return data[path]
            # If not found as key and path looks like a literal value, return it
            try:
                # Try to parse as int first
                if path.isdigit():
                    return path  # Return as string for consistency
                # Try to parse as float
                float(path)
                return path  # Return as string for consistency
            except ValueError:
                pass
            # If it's not found as a key and not a number, still try to access as key
            if isinstance(data, dict):
                return data.get(path)
        
        parts = path.split('.')
        current = data
        
        for part in parts:
            if part == '':
                continue
                
            # Handle array indices like "items[0]"
            if '[' in part and ']' in part:
                array_name, bracket_part = part.split('[', 1)
                index_str = bracket_part.rstrip(']')
                
                if array_name:
                    current = current[array_name]
                
                try:
                    index = int(index_str)
                    current = current[index]
                except (ValueError, IndexError, TypeError):
                    raise ValueError(f"Invalid array index: {index_str}")
            else:
                if isinstance(current, dict):
                    current = current[part]
                else:
                    raise ValueError(f"Cannot access '{part}' on non-dict type")
        
        return current
    
    def render_template(self, data: Dict[str, Any]) -> str:
        """Render HTML using the specified template type."""
        template_map = {
            'key_value': self.render_key_value,
            'title_subtitle_value': self.render_title_subtitle_value,
            'metric_grid': self.render_metric_grid,
            'weather_current': self.render_weather_current,
            'time_display': self.render_time_display,
            'status_list': self.render_status_list,
            'icon_list': self.render_icon_list,
            'text_block': self.render_text_block,
            'chart_simple': self.render_chart_simple,
            'image_caption': self.render_image_caption,
            'rss_headlines': self.render_rss_headlines,
            'rss_summary': self.render_rss_summary,
            'rss_feed_info': self.render_rss_feed_info,
        }
        
        renderer = template_map.get(self.template_type)
        if not renderer:
            return self.render_error(f"Unknown template type: {self.template_type}")
        
        return renderer(data)
    
    def render_key_value(self, data: Dict[str, Any]) -> str:
        """Render key-value template."""
        title = self.safe_str(data.get('title', 'Value'))
        value = self.safe_str(data.get('value', 'N/A'))
        unit = self.safe_str(data.get('unit', ''))
        
        return f"""
        <div class="widget-content">
            <div class="value-display text-center">
                <div class="label mb-sm">{html.escape(title)}</div>
                <div class="value value--large">{html.escape(value)}{html.escape(unit)}</div>
            </div>
        </div>
        """
    
    def render_title_subtitle_value(self, data: Dict[str, Any]) -> str:
        """Render title-subtitle-value template."""
        title = self.safe_str(data.get('title', 'Title'))
        subtitle = self.safe_str(data.get('subtitle', ''))
        value = self.safe_str(data.get('value', 'N/A'))
        description = self.safe_str(data.get('description', ''))
        
        subtitle_html = f'<div class="subtitle mb-sm">{html.escape(subtitle)}</div>' if subtitle else ''
        description_html = f'<div class="description mt-sm">{html.escape(description)}</div>' if description else ''
        
        return f"""
        <div class="widget-content text-center">
            <div class="title mb-md">{html.escape(title)}</div>
            {subtitle_html}
            <div class="value value--large mb-md">{html.escape(value)}</div>
            {description_html}
        </div>
        """
    
    def render_metric_grid(self, data: Dict[str, Any]) -> str:
        """Render metrics grid template."""
        metrics = data.get('metrics', [])
        title_path = data.get('metric_title_path', 'name')
        value_path = data.get('metric_value_path', 'value')
        unit_path = data.get('metric_unit_path', 'unit')
        
        if not isinstance(metrics, list):
            return self.render_error("Metrics must be an array")
        
        metrics_html = []
        for metric in metrics[:8]:  # Limit to 8 metrics for e-paper display
            try:
                metric_title = self.extract_json_path(metric, title_path)
                metric_value = self.extract_json_path(metric, value_path)
                metric_unit = ''
                
                try:
                    metric_unit = self.extract_json_path(metric, unit_path)
                except:
                    pass
                
                metrics_html.append(f"""
                <div class="metric-item">
                    <div class="metric-info">
                        <div class="label">{html.escape(self.safe_str(metric_title))}</div>
                        <div class="value">{html.escape(self.safe_str(metric_value))}{html.escape(self.safe_str(metric_unit))}</div>
                    </div>
                </div>
                """)
            except Exception as e:
                print(f"Warning: Failed to render metric: {e}", file=sys.stderr)
        
        return f"""
        <div class="widget-content">
            <div class="metrics-grid">
                {''.join(metrics_html)}
            </div>
        </div>
        """
    
    def render_weather_current(self, data: Dict[str, Any]) -> str:
        """Render current weather template."""
        temperature = self.safe_str(data.get('temperature', 'N/A'))
        condition = self.safe_str(data.get('condition', 'Unknown'))
        icon = self.safe_str(data.get('icon', '🌤️'))
        humidity = self.safe_str(data.get('humidity', ''))
        wind_speed = self.safe_str(data.get('wind_speed', ''))
        
        details_html = []
        if humidity:
            details_html.append(f'<div class="description">Humidity: {html.escape(humidity)}%</div>')
        if wind_speed:
            details_html.append(f'<div class="description">Wind: {html.escape(wind_speed)} m/s</div>')
        
        return f"""
        <div class="widget-content">
            <div class="weather-current">
                <div class="weather-icon">
                    <div style="font-size: 48px;">{html.escape(icon)}</div>
                </div>
                <div class="weather-temp">
                    <div class="value value--huge">{html.escape(temperature)}°</div>
                    <div class="subtitle">{html.escape(condition)}</div>
                </div>
            </div>
            {''.join(details_html)}
        </div>
        """
    
    def render_time_display(self, data: Dict[str, Any]) -> str:
        """Render time display template."""
        time_val = self.safe_str(data.get('time', datetime.now().strftime('%H:%M:%S')))
        date_val = self.safe_str(data.get('date', datetime.now().strftime('%Y-%m-%d')))
        timezone = self.safe_str(data.get('timezone', ''))
        format_type = self.safe_str(data.get('format', ''))
        
        timezone_html = f'<div class="meta">{html.escape(timezone)}</div>' if timezone else ''
        format_html = f'<div class="meta">{html.escape(format_type)}</div>' if format_type else ''
        
        return f"""
        <div class="widget-content">
            <div class="time-display">
                <div class="time-main">
                    <div class="value value--huge">{html.escape(time_val)}</div>
                    <div class="subtitle">{html.escape(date_val)}</div>
                </div>
                <div class="time-meta">
                    {timezone_html}
                    {format_html}
                </div>
            </div>
        </div>
        """
    
    def render_status_list(self, data: Dict[str, Any]) -> str:
        """Render status list template."""
        items = data.get('items', [])
        name_path = data.get('item_name_path', 'name')
        status_path = data.get('item_status_path', 'status')
        message_path = data.get('item_message_path', 'message')
        
        if not isinstance(items, list):
            return self.render_error("Items must be an array")
        
        items_html = []
        for item in items[:10]:  # Limit to 10 items
            try:
                name = self.extract_json_path(item, name_path)
                status = self.extract_json_path(item, status_path)
                message = ''
                
                try:
                    message = self.extract_json_path(item, message_path)
                except:
                    pass
                
                status_icon = self.get_status_icon(status)
                status_class = self.get_status_class(status)
                
                message_html = f'<div class="description">{html.escape(self.safe_str(message))}</div>' if message else ''
                
                items_html.append(f"""
                <div class="metric-item">
                    <div class="metric-icon">{status_icon}</div>
                    <div class="metric-info">
                        <div class="subtitle">{html.escape(self.safe_str(name))}</div>
                        <div class="status-indicator {status_class}">{html.escape(self.safe_str(status))}</div>
                        {message_html}
                    </div>
                </div>
                """)
            except Exception as e:
                print(f"Warning: Failed to render status item: {e}", file=sys.stderr)
        
        return f"""
        <div class="widget-content">
            <div class="metrics-grid">
                {''.join(items_html)}
            </div>
        </div>
        """
    
    def render_icon_list(self, data: Dict[str, Any]) -> str:
        """Render icon list template."""
        items = data.get('items', [])
        icon_path = data.get('item_icon_path', 'icon')
        title_path = data.get('item_title_path', 'title')
        description_path = data.get('item_description_path', 'description')
        
        if not isinstance(items, list):
            return self.render_error("Items must be an array")
        
        items_html = []
        for item in items[:8]:  # Limit to 8 items
            try:
                icon = ''
                try:
                    icon = self.extract_json_path(item, icon_path)
                except:
                    pass
                
                title = self.extract_json_path(item, title_path)
                description = ''
                
                try:
                    description = self.extract_json_path(item, description_path)
                except:
                    pass
                
                icon_html = f'<div class="metric-icon">{html.escape(self.safe_str(icon))}</div>' if icon else ''
                description_html = f'<div class="description">{html.escape(self.safe_str(description))}</div>' if description else ''
                
                items_html.append(f"""
                <div class="metric-item">
                    {icon_html}
                    <div class="metric-info">
                        <div class="subtitle">{html.escape(self.safe_str(title))}</div>
                        {description_html}
                    </div>
                </div>
                """)
            except Exception as e:
                print(f"Warning: Failed to render icon item: {e}", file=sys.stderr)
        
        return f"""
        <div class="widget-content">
            <div class="metrics-grid">
                {''.join(items_html)}
            </div>
        </div>
        """
    
    def render_text_block(self, data: Dict[str, Any]) -> str:
        """Render text block template."""
        title = self.safe_str(data.get('title', ''))
        content = self.safe_str(data.get('content', 'No content'))
        author = self.safe_str(data.get('author', ''))
        timestamp = self.safe_str(data.get('timestamp', ''))
        
        title_html = f'<div class="title mb-md">{html.escape(title)}</div>' if title else ''
        author_html = f'<div class="meta">— {html.escape(author)}</div>' if author else ''
        timestamp_html = f'<div class="meta">{html.escape(timestamp)}</div>' if timestamp else ''
        
        return f"""
        <div class="widget-content">
            {title_html}
            <div class="description mb-md">{html.escape(content)}</div>
            {author_html}
            {timestamp_html}
        </div>
        """
    
    def render_chart_simple(self, data: Dict[str, Any]) -> str:
        """Render simple chart data template."""
        title = self.safe_str(data.get('title', 'Chart'))
        data_points = data.get('data_points', [])
        labels = data.get('labels', [])
        unit = self.safe_str(data.get('unit', ''))
        
        if not isinstance(data_points, list):
            return self.render_error("Data points must be an array")
        
        # For e-paper, just show key statistics
        if data_points:
            min_val = min(data_points)
            max_val = max(data_points)
            avg_val = sum(data_points) / len(data_points)
            
            stats_html = f"""
            <div class="metrics-grid">
                <div class="metric-item">
                    <div class="metric-info">
                        <div class="label">Min</div>
                        <div class="value">{min_val} {html.escape(unit)}</div>
                    </div>
                </div>
                <div class="metric-item">
                    <div class="metric-info">
                        <div class="label">Max</div>
                        <div class="value">{max_val} {html.escape(unit)}</div>
                    </div>
                </div>
                <div class="metric-item">
                    <div class="metric-info">
                        <div class="label">Avg</div>
                        <div class="value">{avg_val:.1f} {html.escape(unit)}</div>
                    </div>
                </div>
            </div>
            """
        else:
            stats_html = '<div class="description">No data available</div>'
        
        return f"""
        <div class="widget-content">
            <div class="title mb-md">{html.escape(title)}</div>
            {stats_html}
        </div>
        """
    
    def render_image_caption(self, data: Dict[str, Any]) -> str:
        """Render image with caption template."""
        image_url = self.safe_str(data.get('image_url', ''))
        caption = self.safe_str(data.get('caption', ''))
        alt_text = self.safe_str(data.get('alt_text', 'Image'))
        title = self.safe_str(data.get('title', ''))
        
        if not image_url:
            return self.render_error("Image URL is required")
        
        title_html = f'<div class="title mb-md">{html.escape(title)}</div>' if title else ''
        caption_html = f'<div class="description mt-md">{html.escape(caption)}</div>' if caption else ''
        
        return f"""
        <div class="widget-content text-center">
            {title_html}
            <img src="{html.escape(image_url)}" alt="{html.escape(alt_text)}" 
                 style="max-width: 100%; height: auto; border-radius: var(--border-radius);">
            {caption_html}
        </div>
        """
    
    def render_rss_headlines(self, data: Dict[str, Any]) -> str:
        """Render RSS headlines template."""
        feed_title = self.safe_str(data.get('feed_title', 'RSS Feed'))
        items = data.get('items', [])
        
        if not isinstance(items, list):
            return self.render_error("RSS items must be an array")
        
        items_html = []
        for item in items[:10]:  # Limit to 10 headlines
            try:
                if isinstance(item, dict):
                    title = self.safe_str(item.get('title', 'Untitled'))
                    link = self.safe_str(item.get('link', ''))
                    pub_date = self.safe_str(item.get('pub_date', ''))
                    
                    date_html = f'<div class="meta">{html.escape(pub_date)}</div>' if pub_date else ''
                    
                    if link:
                        items_html.append(f"""
                        <div class="metric-item">
                            <div class="metric-info">
                                <div class="subtitle">• <a href="{html.escape(link)}" target="_blank">{html.escape(title)}</a></div>
                                {date_html}
                            </div>
                        </div>
                        """)
                    else:
                        items_html.append(f"""
                        <div class="metric-item">
                            <div class="metric-info">
                                <div class="subtitle">• {html.escape(title)}</div>
                                {date_html}
                            </div>
                        </div>
                        """)
                else:
                    # Handle case where item is just a string (title)
                    items_html.append(f"""
                    <div class="metric-item">
                        <div class="metric-info">
                            <div class="subtitle">• {html.escape(self.safe_str(item))}</div>
                        </div>
                    </div>
                    """)
            except Exception as e:
                print(f"Warning: Failed to render RSS headline: {e}", file=sys.stderr)
        
        if not items_html:
            items_html.append('<div class="description">No headlines available</div>')
        
        return f"""
        <div class="widget-content">
            <div class="title mb-md">{html.escape(feed_title)}</div>
            <div class="metrics-grid">
                {''.join(items_html)}
            </div>
        </div>
        """
    
    def render_rss_summary(self, data: Dict[str, Any]) -> str:
        """Render RSS summary template."""
        feed_title = self.safe_str(data.get('feed_title', 'RSS Feed'))
        items = data.get('items', [])
        
        if not isinstance(items, list) or not items:
            return f"""
            <div class="widget-content text-center">
                <div class="title mb-md">{html.escape(feed_title)}</div>
                <div class="description">No articles available</div>
            </div>
            """
        
        # Get the first item for summary
        item = items[0]
        if isinstance(item, dict):
            title = self.safe_str(item.get('title', 'Untitled'))
            description = self.safe_str(item.get('description', ''))
            author = self.safe_str(item.get('author', ''))
            pub_date = self.safe_str(item.get('pub_date', ''))
            link = self.safe_str(item.get('link', ''))
        else:
            title = self.safe_str(item)
            description = ''
            author = ''
            pub_date = ''
            link = ''
        
        # Truncate description for summary
        if len(description) > 200:
            description = description[:200] + "..."
        
        author_html = f'<div class="meta">By: {html.escape(author)}</div>' if author else ''
        date_html = f'<div class="meta">{html.escape(pub_date)}</div>' if pub_date else ''
        link_html = f'<div class="meta"><a href="{html.escape(link)}" target="_blank">Read more</a></div>' if link else ''
        
        return f"""
        <div class="widget-content">
            <div class="title mb-md">{html.escape(feed_title)}</div>
            <div class="subtitle mb-sm">{html.escape(title)}</div>
            <div class="description mb-md">{html.escape(description)}</div>
            {author_html}
            {date_html}
            {link_html}
        </div>
        """
    
    def render_rss_feed_info(self, data: Dict[str, Any]) -> str:
        """Render RSS feed info template."""
        feed_title = self.safe_str(data.get('feed_title', 'RSS Feed'))
        feed_description = self.safe_str(data.get('feed_description', ''))
        feed_link = self.safe_str(data.get('feed_link', ''))
        item_count = len(data.get('items', []))
        last_updated = self.safe_str(data.get('last_updated', ''))
        
        # Truncate description for display
        if len(feed_description) > 150:
            feed_description = feed_description[:150] + "..."
        
        description_html = f'<div class="description mb-md">{html.escape(feed_description)}</div>' if feed_description else ''
        link_html = f'<div class="meta"><a href="{html.escape(feed_link)}" target="_blank">Visit Feed</a></div>' if feed_link else ''
        updated_html = f'<div class="meta">Last Updated: {html.escape(last_updated)}</div>' if last_updated else ''
        
        return f"""
        <div class="widget-content text-center">
            <div class="title mb-md">{html.escape(feed_title)}</div>
            {description_html}
            <div class="value value--large mb-md">{item_count}</div>
            <div class="subtitle mb-md">Articles Available</div>
            {link_html}
            {updated_html}
        </div>
        """
    
    def render_error(self, message: str) -> str:
        """Render error message."""
        return f"""
        <div class="widget-error">
            <div class="error-icon">⚠️</div>
            <div class="error-message">{html.escape(message)}</div>
        </div>
        """
    
    def safe_str(self, value: Any) -> str:
        """Safely convert value to string."""
        if value is None:
            return ''
        return str(value)
    
    def get_status_icon(self, status: str) -> str:
        """Get status icon based on status value."""
        status_lower = str(status).lower()
        if status_lower in ['online', 'ok', 'healthy', 'success', 'active', 'running']:
            return '✅'
        elif status_lower in ['offline', 'error', 'failed', 'down', 'inactive']:
            return '❌'
        elif status_lower in ['warning', 'degraded', 'partial', 'slow']:
            return '⚠️'
        else:
            return '🔵'
    
    def get_status_class(self, status: str) -> str:
        """Get CSS class based on status value."""
        status_lower = str(status).lower()
        if status_lower in ['online', 'ok', 'healthy', 'success', 'active', 'running']:
            return 'status-indicator--success'
        elif status_lower in ['offline', 'error', 'failed', 'down', 'inactive']:
            return 'status-indicator--error'
        elif status_lower in ['warning', 'degraded', 'partial', 'slow']:
            return 'status-indicator--warning'
        else:
            return 'status-indicator--success'


def main():
    """Main entry point."""
    if len(sys.argv) != 2:
        print("Usage: python3 generic_api_widget.py '<json_config>'", file=sys.stderr)
        sys.exit(1)
    
    try:
        config = json.loads(sys.argv[1])
        executor = WidgetExecutor(config)
        output = executor.execute()
        print(output)
    except json.JSONDecodeError as e:
        print(f"Invalid JSON configuration: {e}", file=sys.stderr)
        sys.exit(1)
    except Exception as e:
        print(f"Widget execution failed: {e}", file=sys.stderr)
        sys.exit(1)


if __name__ == '__main__':
    main()