import { createServer } from 'http';
import { exec } from 'child_process';
import open from '../node_modules/open/index.js';
import path from 'path';
import { fileURLToPath } from 'url';
import fs from 'fs/promises';

// Resolve paths
const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename); // Directory of the current example
const port = 8080;

export default function start(exampleDir) {
    const setupScript = path.resolve(__dirname, './setup-example.sh');

    // Run setup-example.sh
    exec(`${setupScript}`, async (error, stdout, stderr) => {
        if (error) {
            console.error(`Error running setup-example.sh: ${stderr}`);
            process.exit(1);
        }
        console.log(stdout);

        // Start the HTTP server
        const server = startServer(exampleDir);
        if (process.env.OPEN_BROWSER !== 'false') {
            // Wait for the server to start
            await waitForServerReady(`http://localhost:${port}`);
            console.log(`Opening browser at http://localhost:${port}...`);
            // Open the browser after the server starts
            open(`http://localhost:${port}`);
        }

        // Graceful shutdown on process termination
        process.on('SIGINT', () => {
            console.log('\nShutting down server...');
            server.close(() => {
                console.log('Server closed.');
                process.exit(0);
            });
        });
    });
}

// Function to serve static files
function serveStaticFile(res, filePath) {
    fs.readFile(filePath)
        .then((data) => {
            const ext = path.extname(filePath).toLowerCase();
            const contentType = {
                '.html': 'text/html',
                '.js': 'application/javascript',
                '.css': 'text/css',
                '.wasm': 'application/wasm',
            }[ext] || 'application/octet-stream';

            res.writeHead(200, { 'Content-Type': contentType });
            res.end(data);
        })
        .catch((err) => {
            res.writeHead(404, { 'Content-Type': 'text/plain' });
            res.end('404 Not Found');
        });
}

// Create and start the HTTP server
function startServer(exampleDir) {
    const server = createServer((req, res) => {
        const filePath = path.join(exampleDir, req.url === '/' ? '/index.html' : req.url);
        serveStaticFile(res, filePath);
    });

    server.listen(port, () => {
        console.log(`Server running at http://localhost:${port}`);
    });

    return server;
}

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