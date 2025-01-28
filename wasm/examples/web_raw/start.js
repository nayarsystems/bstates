import { fileURLToPath } from 'url';
import path from 'path';
import start from '../start-web-example.js';

// Resolve paths
const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename); // Directory of the current example

// Start web example
start(__dirname);