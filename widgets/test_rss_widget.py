#!/usr/bin/env python3
"""
Test suite for RSS widget functionality in generic_api_widget.py

This test suite covers:
- RSS data fetching and parsing
- RSS configuration handling
- Error scenarios and fallbacks
- Integration with the generic widget executor
"""

import json
import sys
import unittest
from unittest.mock import Mock, patch, MagicMock
import xml.etree.ElementTree as ET

# Import the widget executor
from generic_api_widget import WidgetExecutor


class TestRSSWidget(unittest.TestCase):
    """Test cases for RSS widget functionality."""
    
    def setUp(self):
        """Set up test fixtures."""
        self.mock_rss_xml = '''<?xml version="1.0" encoding="UTF-8"?>
        <rss version="2.0">
            <channel>
                <title>Test RSS Feed</title>
                <description>A test RSS feed</description>
                <link>https://example.com</link>
                <item>
                    <title>Test Article 1</title>
                    <description>This is the first test article</description>
                    <link>https://example.com/article1</link>
                    <author>Test Author</author>
                    <pubDate>Mon, 15 Jan 2024 10:00:00 +0000</pubDate>
                    <guid>article-1</guid>
                </item>
                <item>
                    <title>Test Article 2</title>
                    <description>This is the second test article</description>
                    <link>https://example.com/article2</link>
                    <author>Another Author</author>
                    <pubDate>Mon, 15 Jan 2024 09:00:00 +0000</pubDate>
                    <guid>article-2</guid>
                </item>
            </channel>
        </rss>'''
        
        self.mock_api_response = {
            "feed": {
                "title": "Test RSS Feed",
                "description": "A test RSS feed",
                "link": "https://example.com",
                "items": [
                    {
                        "title": "Test Article 1",
                        "description": "This is the first test article",
                        "link": "https://example.com/article1",
                        "author": "Test Author",
                        "pub_date": "Mon, 15 Jan 2024 10:00:00 +0000",
                        "guid": "article-1"
                    },
                    {
                        "title": "Test Article 2", 
                        "description": "This is the second test article",
                        "link": "https://example.com/article2",
                        "author": "Another Author",
                        "pub_date": "Mon, 15 Jan 2024 09:00:00 +0000",
                        "guid": "article-2"
                    }
                ]
            }
        }
        
        self.rss_config = {
            "data_source": "rss",
            "api_url": "https://example.com/rss",
            "template_type": "rss_headlines",
            "data_mapping": {
                "feed_title": "title",
                "items": "items",
                "max_items": "5"
            },
            "rss_config": {
                "max_items": 5,
                "cache_minutes": 30,
                "include_image": True,
                "include_author": True
            },
            "timeout": 30
        }

    def test_rss_widget_initialization(self):
        """Test RSS widget initialization with proper config."""
        executor = WidgetExecutor(self.rss_config)
        
        self.assertEqual(executor.data_source, "rss")
        self.assertEqual(executor.api_url, "https://example.com/rss")
        self.assertEqual(executor.template_type, "rss_headlines")
        self.assertIsInstance(executor.rss_config, dict)
        self.assertEqual(executor.rss_config["max_items"], 5)

    @patch('requests.post')
    def test_fetch_rss_data_via_api(self, mock_post):
        """Test fetching RSS data via API endpoint."""
        # Mock successful API response
        mock_response = Mock()
        mock_response.raise_for_status.return_value = None
        mock_response.json.return_value = self.mock_api_response
        mock_post.return_value = mock_response
        
        executor = WidgetExecutor(self.rss_config)
        data = executor.fetch_rss_data()
        
        # Verify API was called correctly
        mock_post.assert_called_once()
        call_args = mock_post.call_args
        self.assertEqual(call_args[0][0], "http://localhost:8080/api/rss/preview")
        
        # Verify request payload
        payload = call_args[1]['json']
        self.assertEqual(payload['feed_url'], "https://example.com/rss")
        self.assertEqual(payload['rss_config'], self.rss_config['rss_config'])
        
        # Verify returned data
        self.assertEqual(data['title'], "Test RSS Feed")
        self.assertEqual(len(data['items']), 2)
        self.assertEqual(data['items'][0]['title'], "Test Article 1")

    @patch('requests.post')
    @patch('requests.get')
    def test_fetch_rss_data_fallback_to_direct(self, mock_get, mock_post):
        """Test fallback to direct RSS parsing when API fails."""
        # Mock API failure
        mock_post.side_effect = Exception("Connection error")
        
        # Mock successful direct RSS fetch
        mock_response = Mock()
        mock_response.raise_for_status.return_value = None
        mock_response.content = self.mock_rss_xml.encode('utf-8')
        mock_get.return_value = mock_response
        
        executor = WidgetExecutor(self.rss_config)
        data = executor.fetch_rss_data()
        
        # Verify direct fetch was called
        mock_get.assert_called_once()
        self.assertEqual(mock_get.call_args[0][0], "https://example.com/rss")
        
        # Verify parsed data
        self.assertEqual(data['title'], "Test RSS Feed")
        self.assertEqual(len(data['items']), 2)
        self.assertEqual(data['items'][0]['title'], "Test Article 1")

    def test_get_element_text(self):
        """Test XML element text extraction helper."""
        executor = WidgetExecutor(self.rss_config)
        
        # Parse test XML
        root = ET.fromstring(self.mock_rss_xml)
        channel = root.find('channel')
        
        # Test existing elements
        title = executor.get_element_text(channel, 'title')
        self.assertEqual(title, "Test RSS Feed")
        
        description = executor.get_element_text(channel, 'description')
        self.assertEqual(description, "A test RSS feed")
        
        # Test non-existent element
        missing = executor.get_element_text(channel, 'nonexistent')
        self.assertEqual(missing, "")

    @patch('requests.post')
    def test_rss_data_mapping(self, mock_post):
        """Test data mapping for RSS feeds."""
        # Mock API response
        mock_response = Mock()
        mock_response.raise_for_status.return_value = None
        mock_response.json.return_value = self.mock_api_response
        mock_post.return_value = mock_response
        
        executor = WidgetExecutor(self.rss_config)
        
        # Fetch and map data
        raw_data = executor.fetch_rss_data()
        mapped_data = executor.apply_data_mapping(raw_data)
        
        # Verify mapping
        self.assertEqual(mapped_data['feed_title'], "Test RSS Feed")
        self.assertEqual(len(mapped_data['items']), 2)
        self.assertEqual(mapped_data['max_items'], "5")

    @patch('requests.post')
    def test_rss_widget_execution(self, mock_post):
        """Test complete RSS widget execution."""
        # Mock API response
        mock_response = Mock()
        mock_response.raise_for_status.return_value = None
        mock_response.json.return_value = self.mock_api_response
        mock_post.return_value = mock_response
        
        executor = WidgetExecutor(self.rss_config)
        
        # Execute widget
        result = executor.execute()
        
        # Verify result is HTML string
        self.assertIsInstance(result, str)
        self.assertIn("Test RSS Feed", result)
        self.assertIn("Test Article 1", result)

    def test_rss_widget_error_handling(self):
        """Test error handling for RSS widgets."""
        # Test with missing URL
        config_no_url = self.rss_config.copy()
        config_no_url['api_url'] = ''
        
        executor = WidgetExecutor(config_no_url)
        result = executor.execute()
        
        # Should return error HTML
        self.assertIn("error", result.lower())
        self.assertIn("widget execution failed", result.lower())

    @patch('requests.post')
    def test_rss_api_error_handling(self, mock_post):
        """Test handling of RSS API errors."""
        # Mock API error
        mock_post.side_effect = Exception("API unavailable")
        
        # Also mock direct fetch failure
        with patch('requests.get') as mock_get:
            mock_get.side_effect = Exception("Direct fetch failed")
            
            executor = WidgetExecutor(self.rss_config)
            result = executor.execute()
            
            # Should return error HTML
            self.assertIn("error", result.lower())

    def test_rss_template_types(self):
        """Test different RSS template types."""
        template_configs = [
            ("rss_headlines", "items"),
            ("rss_summary", "items[0].title"),
            ("rss_feed_info", "title")
        ]
        
        for template_type, expected_mapping in template_configs:
            config = self.rss_config.copy()
            config['template_type'] = template_type
            
            executor = WidgetExecutor(config)
            self.assertEqual(executor.template_type, template_type)

    @patch('requests.get')
    def test_direct_rss_xml_parsing(self, mock_get):
        """Test direct RSS XML parsing."""
        # Mock response
        mock_response = Mock()
        mock_response.raise_for_status.return_value = None
        mock_response.content = self.mock_rss_xml.encode('utf-8')
        mock_get.return_value = mock_response
        
        executor = WidgetExecutor(self.rss_config)
        data = executor.fetch_rss_direct()
        
        # Verify parsed data structure
        self.assertIn('title', data)
        self.assertIn('description', data)
        self.assertIn('link', data)
        self.assertIn('items', data)
        
        self.assertEqual(data['title'], "Test RSS Feed")
        self.assertEqual(len(data['items']), 2)
        
        # Verify item structure
        item = data['items'][0]
        self.assertIn('title', item)
        self.assertIn('description', item)
        self.assertIn('link', item)
        self.assertIn('author', item)
        self.assertIn('pub_date', item)

    @patch('requests.get')
    def test_invalid_rss_xml_handling(self, mock_get):
        """Test handling of invalid RSS XML."""
        # Mock response with invalid XML
        invalid_xml = "<?xml version='1.0'?><invalid>not rss</invalid>"
        mock_response = Mock()
        mock_response.raise_for_status.return_value = None
        mock_response.content = invalid_xml.encode('utf-8')
        mock_get.return_value = mock_response
        
        executor = WidgetExecutor(self.rss_config)
        
        with self.assertRaises(Exception) as context:
            executor.fetch_rss_direct()
        
        self.assertIn("no channel element found", str(context.exception))

    def test_rss_config_defaults(self):
        """Test RSS configuration defaults."""
        config = {
            "data_source": "rss",
            "api_url": "https://example.com/rss",
            "template_type": "rss_headlines",
            "data_mapping": {},
        }
        
        executor = WidgetExecutor(config)
        
        # Should have empty RSS config
        self.assertEqual(executor.rss_config, {})
        
        # Fetch should use defaults
        with patch('requests.post') as mock_post:
            mock_response = Mock()
            mock_response.raise_for_status.return_value = None
            mock_response.json.return_value = self.mock_api_response
            mock_post.return_value = mock_response
            
            executor.fetch_rss_data()
            
            # Verify default config was sent
            call_args = mock_post.call_args
            payload = call_args[1]['json']
            self.assertEqual(payload['rss_config'], {})


class TestRSSIntegration(unittest.TestCase):
    """Integration tests for RSS functionality."""

    def test_complete_rss_workflow(self):
        """Test complete RSS widget workflow from config to HTML."""
        config = {
            "data_source": "rss",
            "api_url": "https://feeds.bbci.co.uk/news/rss.xml",  # Real RSS feed
            "template_type": "rss_headlines",
            "data_mapping": {
                "feed_title": "title",
                "items": "items"
            },
            "rss_config": {
                "max_items": 3,
                "cache_minutes": 5
            },
            "timeout": 10
        }
        
        # Mock the API call since we can't make real network calls in tests
        with patch('requests.post') as mock_post:
            mock_response = Mock()
            mock_response.raise_for_status.return_value = None
            mock_response.json.return_value = {
                "feed": {
                    "title": "BBC News",
                    "items": [
                        {"title": "Breaking News 1", "description": "News 1"},
                        {"title": "Breaking News 2", "description": "News 2"},
                        {"title": "Breaking News 3", "description": "News 3"}
                    ]
                }
            }
            mock_post.return_value = mock_response
            
            executor = WidgetExecutor(config)
            result = executor.execute()
            
            # Verify the result contains expected content
            self.assertIsInstance(result, str)
            self.assertIn("BBC News", result)
            self.assertIn("Breaking News", result)


if __name__ == '__main__':
    # Run the tests
    unittest.main(verbosity=2)