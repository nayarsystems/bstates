import puppeteer from 'puppeteer';
import { exec } from 'child_process';
import fetch from 'node-fetch';

let browser;
let page;
let serverProcess;

async function waitForServerReady(url, timeout = 5000) {
    const start = Date.now();
    while (Date.now() - start < timeout) {
        try {
            const response = await fetch(url);
            if (response.ok) return; // Server ready
        } catch {
            // Ignore error and retry
        }
        await new Promise((resolve) => setTimeout(resolve, 100));
    }
    throw new Error('Server did not start');
}

beforeAll(async () => {
    // Start the web server
    serverProcess = exec('OPEN_BROWSER=false node examples/web/start.js');

    // Wait for the server to start
    await waitForServerReady('http://localhost:8080');

    // Launch Puppeteer
    browser = await puppeteer.launch();
    page = await browser.newPage();

    // Navigate to the web application
    await page.goto('http://localhost:8080');

    // Wait for bstates to be loaded
    await page.waitForFunction(() => typeof window.bstates !== 'undefined', {
        timeout: 5000, // Max time to wait in milliseconds
    });
});

afterAll(async () => {
    // Stop the Puppeteer browser
    await browser.close();

    // Stop the web server
    serverProcess.kill('SIGINT');
});

test('Server serves index.html correctly', async () => {
    const response = await fetch('http://localhost:8080');
    const html = await response.text();

    expect(response.status).toBe(200);
    expect(html).toContain('<title>Bstates example</title>'); // Check title in HTML
});

test('Server serves wasm.js correctly', async () => {
    const response = await fetch('http://localhost:8080/main.js');

    expect(response.status).toBe(200);
    const contentType = response.headers.get('content-type');
    expect(contentType).toBe('application/javascript');
});

test('Server returns 404 for missing files', async () => {
    const response = await fetch('http://localhost:8080/nonexistent-file.js');

    expect(response.status).toBe(404);
});

test('Page loads and exposes bstates in the console', async () => {
    // Check the title of the page
    const title = await page.title();
    expect(title).toBe('Bstates example');

    // Check if `bstates` is exposed in the console
    const bstatesExists = await page.evaluate(() => {
        return typeof window.bstates !== 'undefined';
    });
    expect(bstatesExists).toBe(true);
});

test('bstates.createStateQueue works as expected', async () => {
    // Execute code in the browser to validate `createStateQueue`
    const result = await page.evaluate(() => {
        const schema = {
            version: "2.0",
            encoderPipeline: "t:z",
            fields: [
                { name: "field1", type: "int", size: 8 }
            ]
        };

        const queue = window.bstates.createStateQueue(schema).d;
        queue.push({ field1: 123 });
        return queue.toArray();
    });

    expect(result).toEqual([{ field1: 123 }]);
});
