const https = require('https');
const dns = require('dns');

const API_BASE = 'api.minimaxi.com';
const API_PATH = '/v1/chat/completions';

const postData = JSON.stringify({
  model: 'MiniMax-M2.1',
  messages: [
    {
      role: 'user',
      content: '你是编程类模型么？'
    }
  ],
  stream: true
});

console.log('============================================================');
console.log('MiniMax API Direct Call Test');
console.log('============================================================');
console.log('API Endpoint: https://' + API_BASE + API_PATH);
console.log('');

// From DNS resolution start timing
const totalStartTime = Date.now();
console.log('[' + totalStartTime + '] Starting test (DNS resolution)');

// DNS resolution
dns.lookup(API_BASE, (err, address, family) => {
  const dnsTime = Date.now() - totalStartTime;
  console.log('[' + Date.now() + '] DNS resolved: ' + address + ' (IPv' + family + ', ' + dnsTime + 'ms)');

  const options = {
    hostname: API_BASE,
    port: 443,
    path: API_PATH,
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': 'Bearer sk-cp-7p3dBLFjVpvD8px5zWhmnSq1gFu5rHKGEvaS8BSxA0kyJLChaaUs56VNy1-RbmzWdQR8YhHnzDSv36gtzoU2-EifeCAh_o2lEqtzaRjlo88fm0VU2_hlneo',
      'Content-Length': Buffer.byteLength(postData)
    }
  };

  console.log('[' + Date.now() + '] Sending request...');
  console.log('------------------------------------------------------------');

  const req = https.request(options, (res) => {
    const responseStart = Date.now();
    const requestToResponse = responseStart - totalStartTime;

    console.log('[' + Date.now() + '] Response received (time: ' + requestToResponse + 'ms, status: ' + res.statusCode + ')');
    console.log('[' + Date.now() + '] Content-Type: ' + res.headers['content-type']);
    console.log('------------------------------------------------------------');

    let chunkCount = 0;
    let totalBytes = 0;
    const dataChunks = [];
    const chunkTimes = [];

    res.on('data', (chunk) => {
      const now = Date.now();
      const elapsed = now - totalStartTime;
      chunkCount++;
      chunkTimes.push({ num: chunkCount, elapsed: elapsed, bytes: chunk.length });

      console.log('[' + now + '] [elapsed: ' + elapsed + 'ms] Chunk #' + chunkCount + ' (' + chunk.length + ' bytes)');

      dataChunks.push(chunk);
      totalBytes += chunk.length;
    });

    res.on('end', () => {
      const totalTime = Date.now() - totalStartTime;
      console.log('============================================================');
      console.log('Request Complete!');
      console.log('============================================================');
      console.log('Total time: ' + totalTime + 'ms');
      console.log('DNS resolution: ' + dnsTime + 'ms');
      console.log('Request to first response: ' + requestToResponse + 'ms');
      console.log('Total chunks: ' + chunkCount);
      console.log('Total bytes: ' + totalBytes);
      console.log('------------------------------------------------------------');
      console.log('Timeline of chunks:');
      console.log('------------------------------------------------------------');

      // Calculate intervals
      let prevTime = requestToResponse;
      chunkTimes.forEach((item) => {
        const interval = item.elapsed - prevTime;
        console.log('  #' + item.num.toString().padStart(2) + ': +' + interval.toString().padStart(5) + 'ms (' + item.bytes + ' bytes)');
        prevTime = item.elapsed;
      });

      console.log('------------------------------------------------------------');
      console.log('TTFT Analysis:');
      const ttft = chunkTimes[0] ? (chunkTimes[0].elapsed - requestToResponse) : 0;
      console.log('  TTFT (Time To First Token): ' + ttft + 'ms');
      console.log('  TTFT ratio: ' + ((ttft / totalTime) * 100).toFixed(1) + '%');
      console.log('------------------------------------------------------------');

      if (chunkTimes.length > 0) {
        console.log('First chunk preview:');
        console.log(dataChunks[0].toString().substring(0, 500));
        console.log('...');
      }
    });
  });

  req.on('error', (error) => {
    console.error('Request error: ' + error.message);
  });

  req.setTimeout(30000, () => {
    console.error('Request timeout (30s)');
    req.destroy();
  });

  req.write(postData);
  req.end();
  console.log('[' + Date.now() + '] Request sent, waiting for response...');
});
