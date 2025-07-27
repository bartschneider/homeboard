/**
 * E-Paper Dashboard Design System Test Suite
 * Tests the new TRMNL-inspired design system implementation
 */

const { chromium } = require('playwright');

async function testDesignSystem() {
    console.log('🎨 Testing E-Paper Dashboard Design System...\n');
    
    const browser = await chromium.launch({ headless: false });
    const context = await browser.newContext({
        viewport: { width: 1024, height: 768 }
    });
    const page = await context.newPage();

    try {
        // Navigate to dashboard
        console.log('📍 Navigating to http://localhost:8081...');
        await page.goto('http://localhost:8081', { waitUntil: 'networkidle' });
        
        // Test 1: CSS File Loading
        console.log('🔍 Testing CSS file loading...');
        const cssResponse = await page.goto('http://localhost:8081/static/css/design-system.css');
        const cssLoaded = cssResponse.status() === 200;
        console.log(`CSS Status: ${cssResponse.status()} - ${cssLoaded ? '✅ LOADED' : '❌ FAILED'}`);
        
        if (!cssLoaded) {
            console.log('⚠️  CSS file not found - checking static file serving...');
        }
        
        // Go back to dashboard
        await page.goto('http://localhost:8081', { waitUntil: 'networkidle' });
        
        // Test 2: Design System Elements
        console.log('🎯 Testing design system elements...');
        
        const tests = [
            {
                name: 'Dashboard Grid',
                selector: '.dashboard-grid',
                expected: 'Card-based layout grid'
            },
            {
                name: 'Widget Cards',
                selector: '.widget-card',
                expected: 'Individual widget cards'
            },
            {
                name: 'Widget Headers',
                selector: '.widget-header',
                expected: 'Card headers with titles'
            },
            {
                name: 'Status Indicators',
                selector: '.status-indicator',
                expected: 'Execution status badges'
            },
            {
                name: 'Typography Classes',
                selector: '.title, .subtitle, .value, .description',
                expected: 'Design system typography'
            }
        ];
        
        for (const test of tests) {
            const elements = await page.$$(test.selector);
            const found = elements.length > 0;
            console.log(`${test.name}: ${found ? '✅ FOUND' : '❌ MISSING'} (${elements.length} elements)`);
            
            if (found && elements.length > 0) {
                // Get computed styles for first element
                const styles = await page.evaluate((selector) => {
                    const element = document.querySelector(selector);
                    if (!element) return null;
                    
                    const computedStyle = window.getComputedStyle(element);
                    return {
                        display: computedStyle.display,
                        fontFamily: computedStyle.fontFamily,
                        backgroundColor: computedStyle.backgroundColor,
                        border: computedStyle.border,
                        borderRadius: computedStyle.borderRadius,
                        padding: computedStyle.padding,
                        margin: computedStyle.margin
                    };
                }, test.selector);
                
                if (styles) {
                    console.log(`  └─ Styles: display=${styles.display}, border=${styles.border}`);
                }
            }
        }
        
        // Test 3: Console Errors
        console.log('\n🚨 Checking for console errors...');
        const consoleLogs = [];
        page.on('console', msg => {
            consoleLogs.push(`${msg.type()}: ${msg.text()}`);
        });
        
        // Wait a bit to catch any console errors
        await page.waitForTimeout(2000);
        
        const errors = consoleLogs.filter(log => log.startsWith('error'));
        const warnings = consoleLogs.filter(log => log.startsWith('warning'));
        
        console.log(`Console Errors: ${errors.length}`);
        console.log(`Console Warnings: ${warnings.length}`);
        
        if (errors.length > 0) {
            console.log('❌ Console Errors Found:');
            errors.forEach(error => console.log(`  - ${error}`));
        }
        
        // Test 4: Network Requests
        console.log('\n🌐 Analyzing network requests...');
        const responses = [];
        page.on('response', response => {
            responses.push({
                url: response.url(),
                status: response.status(),
                contentType: response.headers()['content-type']
            });
        });
        
        // Reload to capture all network requests
        await page.reload({ waitUntil: 'networkidle' });
        
        const staticRequests = responses.filter(r => r.url.includes('/static/'));
        console.log(`Static file requests: ${staticRequests.length}`);
        
        staticRequests.forEach(req => {
            const status = req.status === 200 ? '✅' : '❌';
            console.log(`  ${status} ${req.status} - ${req.url}`);
        });
        
        // Test 5: Visual Regression Check
        console.log('\n📸 Taking screenshot for visual verification...');
        await page.screenshot({ 
            path: 'dashboard-design-test.png',
            fullPage: true 
        });
        console.log('Screenshot saved: dashboard-design-test.png');
        
        // Test 6: Widget Content Analysis
        console.log('\n🧩 Analyzing widget content...');
        const widgets = await page.$$eval('.widget-card', cards => {
            return cards.map(card => {
                const header = card.querySelector('.widget-header .title');
                const content = card.querySelector('.widget-content');
                const error = card.querySelector('.widget-error');
                
                return {
                    title: header ? header.textContent.trim() : 'No title',
                    hasContent: !!content,
                    hasError: !!error,
                    errorMessage: error ? error.textContent.trim() : null
                };
            });
        });
        
        console.log(`Widgets found: ${widgets.length}`);
        widgets.forEach((widget, i) => {
            const status = widget.hasError ? '❌ ERROR' : '✅ OK';
            console.log(`  ${i + 1}. ${widget.title} - ${status}`);
            if (widget.hasError) {
                console.log(`     Error: ${widget.errorMessage}`);
            }
        });
        
        // Test 7: Responsive Design Check
        console.log('\n📱 Testing responsive design...');
        
        const viewports = [
            { name: 'Desktop', width: 1024, height: 768 },
            { name: 'Tablet', width: 768, height: 1024 },
            { name: 'Mobile', width: 375, height: 667 },
            { name: 'E-Reader', width: 600, height: 800 }
        ];
        
        for (const viewport of viewports) {
            await page.setViewportSize({ width: viewport.width, height: viewport.height });
            await page.waitForTimeout(500);
            
            const gridColumns = await page.evaluate(() => {
                const grid = document.querySelector('.dashboard-grid');
                if (!grid) return 'Grid not found';
                
                const computedStyle = window.getComputedStyle(grid);
                return computedStyle.gridTemplateColumns;
            });
            
            console.log(`${viewport.name} (${viewport.width}x${viewport.height}): ${gridColumns}`);
            
            // Take screenshot for each viewport
            await page.screenshot({ 
                path: `dashboard-${viewport.name.toLowerCase()}-test.png`,
                fullPage: true 
            });
        }
        
        // Summary
        console.log('\n📊 Test Summary:');
        console.log(`CSS File: ${cssLoaded ? '✅ Loaded' : '❌ Missing'}`);
        console.log(`Console Errors: ${errors.length === 0 ? '✅ None' : `❌ ${errors.length} found`}`);
        console.log(`Widgets: ${widgets.length} found`);
        console.log(`Screenshots: 5 captured for analysis`);
        
    } catch (error) {
        console.error('❌ Test execution failed:', error);
    } finally {
        await browser.close();
    }
}

// Run tests
testDesignSystem().catch(console.error);