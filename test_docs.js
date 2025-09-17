const http = require('http');

// Test documentation endpoints
const endpoints = [
    '/api/openapi.json',
    '/docs',
    '/docs-alt',
    '/docs-simple'
];

async function testEndpoint(path) {
    return new Promise((resolve) => {
        const req = http.get({
            hostname: 'localhost',
            port: 3000,
            path: path,
            timeout: 5000
        }, (res) => {
            let data = '';
            res.on('data', chunk => data += chunk);
            res.on('end', () => {
                resolve({
                    path,
                    status: res.statusCode,
                    contentType: res.headers['content-type'],
                    size: data.length,
                    success: res.statusCode === 200
                });
            });
        });

        req.on('error', (err) => {
            resolve({
                path,
                status: 'ERROR',
                error: err.message,
                success: false
            });
        });

        req.on('timeout', () => {
            req.destroy();
            resolve({
                path,
                status: 'TIMEOUT',
                success: false
            });
        });
    });
}

async function testAllEndpoints() {
    console.log('🔍 Testing Kisanlink ERP Documentation Endpoints...\n');

    for (const endpoint of endpoints) {
        const result = await testEndpoint(endpoint);
        const status = result.success ? '✅' : '❌';

        console.log(`${status} ${endpoint}`);
        console.log(`   Status: ${result.status}`);
        if (result.contentType) console.log(`   Content-Type: ${result.contentType}`);
        if (result.size) console.log(`   Size: ${result.size} bytes`);
        if (result.error) console.log(`   Error: ${result.error}`);
        console.log('');
    }

    console.log('📋 Test Summary:');
    console.log('- /api/openapi.json should return JSON specification');
    console.log('- /docs should return Scalar documentation (main)');
    console.log('- /docs-alt should return alternative Scalar documentation');
    console.log('- /docs-simple should return HTML documentation as fallback');
}

testAllEndpoints().catch(console.error);