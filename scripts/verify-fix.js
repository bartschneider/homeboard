const { chromium } = require('playwright');

async function verifyFix() {
    const browser = await chromium.launch({ headless: false });
    const context = await browser.newContext({
        viewport: { width: 1920, height: 1080 }
    });
    const page = await context.newPage();

    console.log('🔍 Verifying dashboard fixes...\n');

    try {
        // Navigate to dashboard
        await page.goto('http://localhost:8081', { 
            waitUntil: 'networkidle',
            timeout: 15000 
        });

        // Take verification screenshot
        await page.screenshot({ path: 'dashboard-fixed.png', fullPage: true });

        // Check if widgets are properly rendered
        const widgetAnalysis = await page.evaluate(() => {
            const widgets = document.querySelectorAll('.widget');
            return Array.from(widgets).map((widget, index) => {
                const content = widget.querySelector('.widget-content');
                const clockWidget = widget.querySelector('.clock-widget');
                const systemWidget = widget.querySelector('.system-widget');
                
                return {
                    index: index + 1,
                    hasClockWidget: !!clockWidget,
                    hasSystemWidget: !!systemWidget,
                    innerHTML: content ? content.innerHTML.substring(0, 150) + '...' : 'No content',
                    hasProperHTML: content && content.innerHTML.includes('<h2>') && !content.innerHTML.includes('&lt;')
                };
            });
        });

        console.log('📊 Widget Rendering Analysis:');
        widgetAnalysis.forEach(widget => {
            console.log(`\n  Widget ${widget.index}:`);
            console.log(`    Clock Widget: ${widget.hasClockWidget ? '✅' : '❌'}`);
            console.log(`    System Widget: ${widget.hasSystemWidget ? '✅' : '❌'}`);
            console.log(`    Proper HTML Rendering: ${widget.hasProperHTML ? '✅' : '❌'}`);
            if (!widget.hasProperHTML) {
                console.log(`    Content Preview: ${widget.innerHTML}`);
            }
        });

        // Check footer rendering
        const footerText = await page.textContent('.footer');
        const hasProperFooter = footerText && !footerText.includes('%!f') && footerText.includes('ms');
        
        console.log(`\n📄 Footer Rendering: ${hasProperFooter ? '✅' : '❌'}`);
        console.log(`Footer Text: ${footerText}`);

        // Check for console errors
        const consoleErrors = [];
        page.on('console', msg => {
            if (msg.type() === 'error') {
                consoleErrors.push(msg.text());
            }
        });

        // Test favicon
        try {
            const faviconResponse = await page.goto('http://localhost:8081/favicon.ico');
            const faviconStatus = faviconResponse?.status();
            console.log(`\n🖼️  Favicon: ${faviconStatus === 200 ? '✅' : '❌'} (Status: ${faviconStatus})`);
        } catch (e) {
            console.log(`\n🖼️  Favicon: ❌ (Error: ${e.message})`);
        }

        // Final summary
        const allWidgetsWorking = widgetAnalysis.every(w => w.hasProperHTML);
        const clockWorking = widgetAnalysis.some(w => w.hasClockWidget);
        const systemWidgetPresent = widgetAnalysis.some(w => w.hasSystemWidget);

        console.log('\n🎯 VERIFICATION SUMMARY');
        console.log('='.repeat(30));
        console.log(`HTML Rendering Fixed: ${allWidgetsWorking ? '✅ YES' : '❌ NO'}`);
        console.log(`Clock Widget Working: ${clockWorking ? '✅ YES' : '❌ NO'}`);
        console.log(`System Widget Present: ${systemWidgetPresent ? '✅ YES' : '❌ NO'}`);
        console.log(`Footer Fixed: ${hasProperFooter ? '✅ YES' : '❌ NO'}`);
        console.log(`Console Errors: ${consoleErrors.length === 0 ? '✅ NONE' : `❌ ${consoleErrors.length}`}`);

        if (allWidgetsWorking && clockWorking && hasProperFooter) {
            console.log('\n🎉 SUCCESS: Dashboard is now fully functional!');
        } else {
            console.log('\n⚠️  PARTIAL: Some issues remain to be addressed.');
        }

        console.log('\n📸 Screenshot saved as: dashboard-fixed.png');

    } catch (error) {
        console.error(`❌ Verification failed: ${error.message}`);
        await page.screenshot({ path: 'verification-error.png', fullPage: true });
    }

    await browser.close();
}

// Run verification
verifyFix().catch(console.error);