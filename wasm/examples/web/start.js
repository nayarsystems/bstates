import { createServer } from 'http';
import { exec } from 'child_process';
import open from 'open';
import path from 'path';
import { fileURLToPath } from 'url';
import fs from 'fs/promises';

// Resolve paths
const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const setupScript = path.resolve(__dirname, '../setup-example.sh');
const exampleDir = __dirname; // Directory of the current example
const port = 8080;

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
function startServer() {
    const server = createServer((req, res) => {
        const filePath = path.join(exampleDir, req.url === '/' ? '/index.html' : req.url);
        serveStaticFile(res, filePath);
    });

    server.listen(port, () => {
        console.log(`Server running at http://localhost:${port}`);
    });

    return server;
}

// Run setup-example.sh
exec(`${setupScript}`, (error, stdout, stderr) => {
    if (error) {
        console.error(`Error running setup-example.sh: ${stderr}`);
        process.exit(1);
    }
    console.log(stdout);

    // Start the HTTP server
    const server = startServer();

    // Open the browser after the server starts
    setTimeout(() => {
        console.log(`Opening browser at http://localhost:${port}...`);
        open(`http://localhost:${port}`);
    }, 2000);

    // Graceful shutdown on process termination
    process.on('SIGINT', () => {
        console.log('\nShutting down server...');
        server.close(() => {
            console.log('Server closed.');
            process.exit(0);
        });
    });
});
