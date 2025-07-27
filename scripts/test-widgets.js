const { chromium } = require('playwright');

async function testWidgets() {
    const browser = await chromium.launch({ headless: false });
    const context = await browser.newContext({
        viewport: { width: 1920, height: 1080 }
    });
    const page = await context.newPage();

    // Set up comprehensive logging
    const consoleMessages = [];
    const networkRequests = [];
    const failedRequests = [];

    page.on('console', msg => {
        consoleMessages.push({
            type: msg.type(),
            text: msg.text(),
            location: msg.location(),
            timestamp: new Date().toISOString()
        });
        console.log(`[${new Date().toISOString()}] [CONSOLE ${msg.type().toUpperCase()}] ${msg.text()}`);
    });

    page.on('request', request => {
        networkRequests.push({
            url: request.url(),
            method: request.method(),
            timestamp: new Date().toISOString()
        });
    });

    page.on('requestfailed', request => {
        failedRequests.push({
            url: request.url(),
            failure: request.failure(),
            method: request.method(),
            timestamp: new Date().toISOString()
        });
        console.log(`[${new Date().toISOString()}] [NETWORK FAILED] ${request.method()} ${request.url()} - ${request.failure()?.errorText}`);
    });

    console.log('🧪 Starting comprehensive widget testing...\n');

    try {
        // 1. Navigate to dashboard
        console.log('📡 Loading dashboard...');
        await page.goto('http://localhost:8081', { 
            waitUntil: 'networkidle',
            timeout: 30000 
        });

        // 2. Wait for page to fully load
        await page.waitForTimeout(3000);

        // 3. Take initial screenshot
        await page.screenshot({ path: 'widget-test-initial.png', fullPage: true });
        console.log('📸 Initial screenshot saved\n');

        // 4. Analyze page structure
        console.log('🔍 Analyzing page structure...');
        
        const pageAnalysis = await page.evaluate(() => {
            return {
                title: document.title,
                hasHeader: !!document.querySelector('.header'),
                hasWidgetContainer: !!document.querySelector('.widgets-container'),
                widgets: Array.from(document.querySelectorAll('.widget')).map((widget, index) => {
                    const content = widget.querySelector('.widget-content');
                    const error = widget.querySelector('.widget-error');
                    return {
                        index: index,
                        hasContent: !!content,
                        hasError: !!error,
                        innerHTML: widget.innerHTML.substring(0, 200) + '...',
                        textContent: widget.textContent.substring(0, 100) + '...',
                        errorMessage: error ? error.textContent : null
                    };
                }),
                footerText: document.querySelector('.footer')?.textContent || 'No footer found',
                bodyHTML: document.body.innerHTML.length,
                scripts: document.scripts.length,
                stylesheets: document.styleSheets.length
            };
        });

        console.log(`Page Title: "${pageAnalysis.title}"`);
        console.log(`Header Present: ${pageAnalysis.hasHeader ? '✅' : '❌'}`);
        console.log(`Widget Container: ${pageAnalysis.hasWidgetContainer ? '✅' : '❌'}`);
        console.log(`Widgets Found: ${pageAnalysis.widgets.length}`);
        console.log(`Scripts Loaded: ${pageAnalysis.scripts}`);
        console.log(`Stylesheets: ${pageAnalysis.stylesheets}`);
        console.log(`Body HTML Size: ${pageAnalysis.bodyHTML} characters`);
        console.log(`Footer: ${pageAnalysis.footerText.substring(0, 50)}...`);

        console.log('\n📋 Widget Analysis:');
        pageAnalysis.widgets.forEach((widget, i) => {
            console.log(`\n  Widget ${i + 1}:`);
            console.log(`    Has Content: ${widget.hasContent ? '✅' : '❌'}`);
            console.log(`    Has Error: ${widget.hasError ? '❌' : '✅'}`);
            if (widget.hasError && widget.errorMessage) {
                console.log(`    Error Message: ${widget.errorMessage.substring(0, 100)}...`);
            }
            console.log(`    Text Preview: ${widget.textContent.substring(0, 60)}...`);
        });

        // 5. Test API endpoints
        console.log('\n🌐 Testing API endpoints...');
        
        const apiTests = [
            { name: 'Health Check', url: '/health' },
            { name: 'Config', url: '/api/config' }
        ];

        for (const test of apiTests) {
            try {
                const response = await page.goto(`http://localhost:8081${test.url}`, { timeout: 10000 });
                const status = response?.status();
                const contentType = response?.headers()['content-type'] || '';
                
                console.log(`  ${test.name} (${test.url}): ${status} (${contentType})`);
                
                if (status === 200) {
                    const text = await response.text();
                    console.log(`    Response: ${text.substring(0, 100)}${text.length > 100 ? '...' : ''}`);
                }
            } catch (e) {
                console.log(`  ${test.name} (${test.url}): ❌ Failed - ${e.message}`);
            }
        }

        // Return to dashboard
        await page.goto('http://localhost:8081');

        // 6. Test widget refresh
        console.log('\n🔄 Testing widget refresh...');
        try {
            // Simulate key press for manual refresh
            await page.keyboard.press('r');
            await page.waitForTimeout(3000);
            console.log('✅ Manual refresh test completed');
            
            // Take screenshot after refresh
            await page.screenshot({ path: 'widget-test-after-refresh.png', fullPage: true });
            console.log('📸 Post-refresh screenshot saved');
        } catch (e) {
            console.log(`❌ Manual refresh failed: ${e.message}`);
        }

        // 7. Performance analysis
        console.log('\n⚡ Performance Analysis...');
        const perfMetrics = await page.evaluate(() => {
            const perf = performance.getEntriesByType('navigation')[0];
            const resources = performance.getEntriesByType('resource');
            
            return {
                navigation: {
                    domContentLoaded: Math.round(perf.domContentLoadedEventEnd - perf.domContentLoadedEventStart),
                    loadComplete: Math.round(perf.loadEventEnd - perf.loadEventStart),
                    totalTime: Math.round(perf.loadEventEnd - perf.navigationStart)
                },
                resources: {
                    total: resources.length,
                    scripts: resources.filter(r => r.name.includes('.js')).length,
                    styles: resources.filter(r => r.name.includes('.css')).length,
                    images: resources.filter(r => r.name.includes('.png') || r.name.includes('.jpg')).length
                },
                memory: performance.memory ? {
                    used: Math.round(performance.memory.usedJSHeapSize / 1024 / 1024),
                    total: Math.round(performance.memory.totalJSHeapSize / 1024 / 1024),
                    limit: Math.round(performance.memory.jsHeapSizeLimit / 1024 / 1024)
                } : null
            };
        });

        console.log(`DOM Content Loaded: ${perfMetrics.navigation.domContentLoaded}ms`);
        console.log(`Load Complete: ${perfMetrics.navigation.loadComplete}ms`);
        console.log(`Total Load Time: ${perfMetrics.navigation.totalTime}ms`);
        console.log(`Resources Loaded: ${perfMetrics.resources.total} (${perfMetrics.resources.scripts} scripts, ${perfMetrics.resources.styles} styles)`);
        
        if (perfMetrics.memory) {
            console.log(`Memory Usage: ${perfMetrics.memory.used}MB / ${perfMetrics.memory.limit}MB`);
        }

        // 8. Test admin panel
        console.log('\n👤 Testing admin panel...');
        try {
            await page.goto('http://localhost:8081/admin', { timeout: 10000 });
            
            const adminAnalysis = await page.evaluate(() => {
                return {
                    title: document.title,
                    hasForm: !!document.querySelector('form'),
                    formElements: document.querySelectorAll('input, select, textarea').length,
                    buttons: document.querySelectorAll('button, input[type="submit"]').length,
                    textContent: document.body.textContent.length
                };
            });

            console.log(`Admin Title: "${adminAnalysis.title}"`);
            console.log(`Has Form: ${adminAnalysis.hasForm ? '✅' : '❌'}`);
            console.log(`Form Elements: ${adminAnalysis.formElements}`);
            console.log(`Buttons: ${adminAnalysis.buttons}`);
            console.log(`Content Length: ${adminAnalysis.textContent} characters`);

            await page.screenshot({ path: 'widget-test-admin.png', fullPage: true });
            console.log('📸 Admin panel screenshot saved');

        } catch (e) {
            console.log(`❌ Admin panel test failed: ${e.message}`);
        }

        // 9. Final diagnostic
        console.log('\n🔍 Final Diagnostic Summary...');
        console.log('='.repeat(50));

        // Categorize console messages
        const errors = consoleMessages.filter(m => m.type === 'error');
        const warnings = consoleMessages.filter(m => m.type === 'warning');
        const logs = consoleMessages.filter(m => m.type === 'log');

        console.log(`Console Messages: ${consoleMessages.length} total`);
        console.log(`  Errors: ${errors.length}`);
        console.log(`  Warnings: ${warnings.length}`);
        console.log(`  Logs: ${logs.length}`);

        if (errors.length > 0) {
            console.log('\n❌ ERRORS FOUND:');
            errors.forEach((error, i) => {
                console.log(`  ${i + 1}. ${error.text}`);
                if (error.location) {
                    console.log(`     Location: ${error.location.url}:${error.location.lineNumber}`);
                }
            });
        }

        console.log(`\nNetwork Requests: ${networkRequests.length} total`);
        console.log(`Failed Requests: ${failedRequests.length}`);

        if (failedRequests.length > 0) {
            console.log('\n🌐 NETWORK FAILURES:');
            failedRequests.forEach((req, i) => {
                console.log(`  ${i + 1}. ${req.method} ${req.url}`);
                console.log(`     Error: ${req.failure?.errorText || 'Unknown'}`);
            });
        }

        console.log('\n📁 Generated Screenshots:');
        console.log('  - widget-test-initial.png');
        console.log('  - widget-test-after-refresh.png');
        console.log('  - widget-test-admin.png');

    } catch (error) {
        console.error(`❌ Test failed: ${error.message}`);
        await page.screenshot({ path: 'widget-test-error.png', fullPage: true });
        console.log('📸 Error screenshot saved as widget-test-error.png');
    }

    await browser.close();
    console.log('\n✅ Widget testing complete!');
}

// Run the test
testWidgets().catch(console.error);