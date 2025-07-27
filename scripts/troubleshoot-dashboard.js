const { chromium } = require('playwright');

async function troubleshootDashboard() {
    const browser = await chromium.launch({ headless: false });
    const context = await browser.newContext({
        viewport: { width: 1920, height: 1080 }
    });
    const page = await context.newPage();

    // Set up console logging
    const consoleMessages = [];
    page.on('console', msg => {
        consoleMessages.push({
            type: msg.type(),
            text: msg.text(),
            location: msg.location()
        });
        console.log(`[CONSOLE ${msg.type().toUpperCase()}] ${msg.text()}`);
    });

    // Set up network monitoring
    const networkRequests = [];
    const failedRequests = [];
    
    page.on('request', request => {
        networkRequests.push({
            url: request.url(),
            method: request.method(),
            timestamp: Date.now()
        });
    });

    page.on('requestfailed', request => {
        failedRequests.push({
            url: request.url(),
            failure: request.failure(),
            method: request.method()
        });
        console.log(`[NETWORK FAILED] ${request.method()} ${request.url()} - ${request.failure()?.errorText}`);
    });

    page.on('response', response => {
        if (!response.ok()) {
            console.log(`[HTTP ERROR] ${response.status()} ${response.url()}`);
        }
    });

    console.log('🔍 Starting E-Paper Dashboard troubleshooting...\n');

    try {
        // 1. Navigate to the dashboard
        console.log('📡 Navigating to http://localhost:8081...');
        const startTime = Date.now();
        await page.goto('http://localhost:8081', { 
            waitUntil: 'networkidle',
            timeout: 30000 
        });
        const loadTime = Date.now() - startTime;
        console.log(`✅ Page loaded in ${loadTime}ms\n`);

        // 2. Take initial screenshot
        await page.screenshot({ path: 'dashboard-initial.png', fullPage: true });
        console.log('📸 Initial screenshot saved as dashboard-initial.png\n');

        // 3. Check page title and basic structure
        const title = await page.title();
        console.log(`📄 Page title: "${title}"`);
        
        const bodyText = await page.textContent('body');
        const hasContent = bodyText && bodyText.trim().length > 100;
        console.log(`📝 Body content loaded: ${hasContent ? '✅ Yes' : '❌ No'} (${bodyText?.length || 0} chars)\n`);

        // 4. Check for widgets
        console.log('🧩 Checking for widgets...');
        const widgets = await page.$$('[class*="widget"], [data-widget], .dashboard-widget, .card, .panel');
        console.log(`Found ${widgets.length} potential widget elements`);

        for (let i = 0; i < Math.min(widgets.length, 5); i++) {
            try {
                const widget = widgets[i];
                const className = await widget.getAttribute('class');
                const id = await widget.getAttribute('id');
                const text = await widget.textContent();
                console.log(`  Widget ${i + 1}: class="${className}" id="${id}" text="${text?.substring(0, 50)}..."`);
            } catch (e) {
                console.log(`  Widget ${i + 1}: Error reading widget - ${e.message}`);
            }
        }
        console.log();

        // 5. Check JavaScript execution
        console.log('⚙️ Testing JavaScript execution...');
        try {
            const jsTest = await page.evaluate(() => {
                return {
                    jquery: typeof $ !== 'undefined',
                    windowLoaded: document.readyState === 'complete',
                    scripts: document.scripts.length,
                    hasErrors: window.onerror !== null
                };
            });
            console.log(`jQuery available: ${jsTest.jquery ? '✅' : '❌'}`);
            console.log(`Document ready: ${jsTest.windowLoaded ? '✅' : '❌'}`);
            console.log(`Script tags found: ${jsTest.scripts}`);
        } catch (e) {
            console.log(`❌ JavaScript execution error: ${e.message}`);
        }
        console.log();

        // 6. Check API endpoints
        console.log('🌐 Testing API endpoints...');
        const apiEndpoints = [
            '/api/widgets',
            '/api/config',
            '/api/weather',
            '/api/system',
            '/api/status'
        ];

        for (const endpoint of apiEndpoints) {
            try {
                const response = await page.goto(`http://localhost:8081${endpoint}`, { timeout: 5000 });
                const status = response?.status();
                const contentType = response?.headers()['content-type'] || '';
                console.log(`  ${endpoint}: ${status} (${contentType})`);
                
                if (status === 200 && contentType.includes('json')) {
                    const text = await response.text();
                    console.log(`    Response: ${text.substring(0, 100)}...`);
                }
            } catch (e) {
                console.log(`  ${endpoint}: ❌ Failed - ${e.message}`);
            }
        }

        // Return to main page
        await page.goto('http://localhost:8081');
        console.log();

        // 7. Check admin panel
        console.log('👤 Testing admin panel at /admin...');
        try {
            await page.goto('http://localhost:8081/admin', { timeout: 10000 });
            await page.screenshot({ path: 'admin-panel.png', fullPage: true });
            const adminTitle = await page.title();
            const adminContent = await page.textContent('body');
            console.log(`Admin page title: "${adminTitle}"`);
            console.log(`Admin content loaded: ${adminContent?.length > 50 ? '✅ Yes' : '❌ No'}`);
            console.log('📸 Admin panel screenshot saved as admin-panel.png');
        } catch (e) {
            console.log(`❌ Admin panel error: ${e.message}`);
        }
        console.log();

        // 8. Performance metrics
        console.log('📊 Collecting performance metrics...');
        const metrics = await page.evaluate(() => {
            const perf = performance.getEntriesByType('navigation')[0];
            return {
                domContentLoaded: perf.domContentLoadedEventEnd - perf.domContentLoadedEventStart,
                loadComplete: perf.loadEventEnd - perf.loadEventStart,
                totalTime: perf.loadEventEnd - perf.navigationStart,
                resourceCount: performance.getEntriesByType('resource').length
            };
        });
        
        console.log(`DOM Content Loaded: ${metrics.domContentLoaded}ms`);
        console.log(`Load Event: ${metrics.loadComplete}ms`);
        console.log(`Total Load Time: ${metrics.totalTime}ms`);
        console.log(`Resources Loaded: ${metrics.resourceCount}`);
        console.log();

        // 9. Widget functionality test
        console.log('🔧 Testing widget functionality...');
        await page.goto('http://localhost:8081');
        
        // Look for interactive elements
        const buttons = await page.$$('button, .btn, [role="button"]');
        const inputs = await page.$$('input, select, textarea');
        const links = await page.$$('a[href]');
        
        console.log(`Interactive elements found:`);
        console.log(`  Buttons: ${buttons.length}`);
        console.log(`  Inputs: ${inputs.length}`);
        console.log(`  Links: ${links.length}`);

        // Try to interact with widgets
        if (buttons.length > 0) {
            try {
                console.log('🖱️  Testing button interaction...');
                await buttons[0].click();
                await page.waitForTimeout(2000);
                console.log('✅ Button click successful');
            } catch (e) {
                console.log(`❌ Button interaction failed: ${e.message}`);
            }
        }

        // 10. Final screenshot
        await page.screenshot({ path: 'dashboard-final.png', fullPage: true });
        console.log('📸 Final screenshot saved as dashboard-final.png\n');

    } catch (error) {
        console.error(`❌ Critical error during troubleshooting: ${error.message}`);
        await page.screenshot({ path: 'error-screenshot.png', fullPage: true });
    }

    // Generate summary report
    console.log('📋 TROUBLESHOOTING SUMMARY\n');
    console.log('='.repeat(50));
    
    console.log(`\n🔍 Console Messages (${consoleMessages.length} total):`);
    const errorMessages = consoleMessages.filter(msg => msg.type === 'error');
    const warningMessages = consoleMessages.filter(msg => msg.type === 'warning');
    
    if (errorMessages.length > 0) {
        console.log(`❌ Errors (${errorMessages.length}):`);
        errorMessages.forEach((msg, i) => {
            console.log(`  ${i + 1}. ${msg.text}`);
            if (msg.location) {
                console.log(`     Location: ${msg.location.url}:${msg.location.lineNumber}`);
            }
        });
    }
    
    if (warningMessages.length > 0) {
        console.log(`⚠️  Warnings (${warningMessages.length}):`);
        warningMessages.slice(0, 5).forEach((msg, i) => {
            console.log(`  ${i + 1}. ${msg.text}`);
        });
    }

    console.log(`\n🌐 Network Issues (${failedRequests.length} failed requests):`);
    if (failedRequests.length > 0) {
        failedRequests.forEach((req, i) => {
            console.log(`  ${i + 1}. ${req.method} ${req.url}`);
            console.log(`     Error: ${req.failure?.errorText || 'Unknown error'}`);
        });
    } else {
        console.log('✅ No failed network requests detected');
    }

    console.log(`\n📊 Total Network Requests: ${networkRequests.length}`);
    
    console.log('\n📁 Generated Files:');
    console.log('  - dashboard-initial.png (initial state)');
    console.log('  - admin-panel.png (admin interface)');
    console.log('  - dashboard-final.png (final state)');
    console.log('  - error-screenshot.png (if errors occurred)');

    await browser.close();
    console.log('\n✅ Troubleshooting complete!');
}

// Run the troubleshooting
troubleshootDashboard().catch(console.error);