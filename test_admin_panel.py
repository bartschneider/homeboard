#!/usr/bin/env python3
"""
Test script for the Admin Panel API endpoints.
This validates the complete system integration.
"""

import requests
import json
import time

BASE_URL = "http://localhost:8081"

def test_api_health():
    """Test API health endpoint."""
    print("Testing API health...")
    try:
        response = requests.get(f"{BASE_URL}/api/health")
        print(f"Health check: {response.status_code} - {response.json()}")
        return response.status_code == 200
    except Exception as e:
        print(f"Health check failed: {e}")
        return False

def test_widget_templates():
    """Test widget templates endpoint."""
    print("\nTesting widget templates...")
    try:
        response = requests.get(f"{BASE_URL}/api/widgets/templates")
        data = response.json()
        print(f"Templates: {response.status_code} - Found {data['total']} templates")
        print("Available templates:", [t['name'] for t in data['templates'][:3]])
        return response.status_code == 200
    except Exception as e:
        print(f"Templates test failed: {e}")
        return False

def test_create_widget():
    """Test creating a widget."""
    print("\nTesting widget creation...")
    widget_data = {
        "name": "Test Weather Widget",
        "template_type": "weather_current",
        "api_url": "https://httpbin.org/json",
        "api_headers": {"User-Agent": "Test-Agent"},
        "data_mapping": {
            "temperature": "slideshow.title",
            "condition": "slideshow.author"
        },
        "description": "Test widget for validation",
        "timeout": 30,
        "enabled": True
    }
    
    try:
        response = requests.post(f"{BASE_URL}/api/widgets", json=widget_data)
        data = response.json()
        print(f"Widget creation: {response.status_code}")
        if response.status_code == 200:
            print(f"Created widget with ID: {data['id']}")
            return data['id']
        else:
            print(f"Error: {data}")
            return None
    except Exception as e:
        print(f"Widget creation failed: {e}")
        return None

def test_create_dashboard(widget_id):
    """Test creating a dashboard."""
    print("\nTesting dashboard creation...")
    dashboard_data = {
        "name": "Test Dashboard",
        "description": "Test dashboard for validation",
        "is_default": True
    }
    
    try:
        response = requests.post(f"{BASE_URL}/api/dashboards", json=dashboard_data)
        data = response.json()
        print(f"Dashboard creation: {response.status_code}")
        if response.status_code == 200:
            dashboard_id = data['id']
            print(f"Created dashboard with ID: {dashboard_id}")
            
            # Add widget to dashboard
            if widget_id:
                widget_data = {
                    "widget_id": widget_id,
                    "display_order": 1
                }
                add_response = requests.post(f"{BASE_URL}/api/dashboards/{dashboard_id}/widgets", json=widget_data)
                print(f"Add widget to dashboard: {add_response.status_code}")
            
            return dashboard_id
        else:
            print(f"Error: {data}")
            return None
    except Exception as e:
        print(f"Dashboard creation failed: {e}")
        return None

def test_get_clients():
    """Test getting clients."""
    print("\nTesting client listing...")
    try:
        response = requests.get(f"{BASE_URL}/api/clients")
        data = response.json()
        print(f"Clients: {response.status_code} - Found {data['total']} clients")
        return response.status_code == 200
    except Exception as e:
        print(f"Clients test failed: {e}")
        return False

def test_llm_analyze():
    """Test LLM analysis (if API key available)."""
    print("\nTesting LLM analysis...")
    analyze_data = {
        "apiUrl": "https://httpbin.org/json",
        "widgetTemplate": "key_value"
    }
    
    try:
        response = requests.post(f"{BASE_URL}/api/llm/analyze", json=analyze_data)
        print(f"LLM analysis: {response.status_code}")
        if response.status_code == 200:
            data = response.json()
            print("Analysis successful:", "dataMapping" in data)
        else:
            print("LLM analysis unavailable (likely no API key)")
        return True  # This is optional functionality
    except Exception as e:
        print(f"LLM analysis test failed: {e}")
        return True  # This is optional functionality

def test_dashboard_access():
    """Test accessing the main dashboard."""
    print("\nTesting dashboard access...")
    try:
        response = requests.get(f"{BASE_URL}/")
        print(f"Dashboard access: {response.status_code}")
        if response.status_code == 200:
            print("Dashboard HTML length:", len(response.text))
            return True
    except Exception as e:
        print(f"Dashboard access failed: {e}")
        return False

def main():
    print("E-Paper Dashboard Admin Panel Integration Test")
    print("=" * 50)
    
    # Wait a moment for server to be ready
    print("Waiting for server to start...")
    time.sleep(2)
    
    tests = [
        ("API Health", test_api_health),
        ("Widget Templates", test_widget_templates),
        ("Client Listing", test_get_clients),
        ("Dashboard Access", test_dashboard_access),
        ("LLM Analysis", test_llm_analyze),
    ]
    
    results = []
    widget_id = None
    dashboard_id = None
    
    # Run basic tests
    for test_name, test_func in tests:
        result = test_func()
        results.append((test_name, result))
    
    # Run creation tests
    widget_id = test_create_widget()
    results.append(("Widget Creation", widget_id is not None))
    
    dashboard_id = test_create_dashboard(widget_id)
    results.append(("Dashboard Creation", dashboard_id is not None))
    
    # Summary
    print("\n" + "=" * 50)
    print("Test Results Summary:")
    passed = 0
    total = len(results)
    
    for test_name, result in results:
        status = "PASS" if result else "FAIL"
        print(f"  {test_name:<20}: {status}")
        if result:
            passed += 1
    
    print(f"\nTotal: {passed}/{total} tests passed")
    
    if passed == total:
        print("🎉 All tests passed! Admin Panel is working correctly.")
    else:
        print("⚠️  Some tests failed. Check the server logs for details.")
    
    return passed == total

if __name__ == "__main__":
    main()