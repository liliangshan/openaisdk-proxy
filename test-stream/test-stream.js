const http = require('http');
const dns = require('dns');

const API_HOST = '127.0.0.1';
const API_PORT = 28080;
const API_PATH = '/api/v1/chat/completions';

const postData = JSON.stringify({
  model: 'mm-MiniMax-M2.1',
  messages: [
    {
      role: 'system',
      content: '你是一个专业、友好的AI编程助手，善于解决各种编程问题。'
    },
    {
      role: 'user',
      content: '你是编程类模型么？'
    }
  ],
  reasoning: {
    enabled: true
  },
  stream: true
});

console.log('============================================================');
console.log('流式输出解析测试');
console.log('============================================================');

// 从 DNS 解析开始计时
const totalStartTime = Date.now();
console.log('[' + totalStartTime + '] 开始测试 (DNS 解析)');

// DNS 解析
dns.lookup(API_HOST, (err, address, family) => {
  const dnsTime = Date.now() - totalStartTime;
  console.log('[' + Date.now() + '] DNS 解析完成: ' + address + ' (IPv' + family + ', ' + dnsTime + 'ms)');

  // 创建 TCP 连接
  const connectStart = Date.now();
  const socket = new net.Socket();

  socket.connect(API_PORT, API_HOST, () => {
    const tcpTime = Date.now() - connectStart;
    console.log('[' + Date.now() + '] TCP 连接建立 (耗时: ' + tcpTime + 'ms)');

    // 创建 HTTP 请求
    const options = {
      hostname: API_HOST,
      port: API_PORT,
      path: API_PATH,
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer sk_ca480b06bbd78b4a816617a29a5b7f44c9dce6438cbe192bc8c4d2828cd2d99c',
        'Content-Length': Buffer.byteLength(postData)
      }
    };

    const req = http.request(options, (res) => {
      const responseStart = Date.now();
      const requestToResponse = responseStart - totalStartTime;

      console.log('[' + Date.now() + '] 收到响应头 (耗时: ' + requestToResponse + 'ms, 状态码: ' + res.statusCode + ')');
      console.log('------------------------------------------------------------');

      let chunkCount = 0;
      let totalBytes = 0;
      const chunkTimes = [];
      let fullContent = '';
      let reasoningContent = '';

      // 解析 SSE 数据块
      const parseSSE = (data) => {
        const lines = data.split('\n');
        let content = '';
        let reasoning = '';

        for (const line of lines) {
          if (line.startsWith('data: ')) {
            const jsonStr = line.slice(6).trim();
            if (jsonStr === '') continue;

            try {
              const parsed = JSON.parse(jsonStr);
              if (parsed.choices && parsed.choices[0]) {
                const delta = parsed.choices[0].delta;
                if (delta.content) {
                  content += delta.content;
                }
              }
            } catch (e) {
              // 解析失败，忽略
            }
          }
        }
        return content;
      };

      res.on('data', (chunk) => {
        const now = Date.now();
        const elapsed = now - totalStartTime;
        chunkCount++;
        chunkTimes.push({ num: chunkCount, elapsed: elapsed, bytes: chunk.length });

        const chunkStr = chunk.toString();
        console.log('[' + now + '] [' + elapsed + 'ms] Chunk #' + chunkCount + ' (' + chunk.length + ' bytes)');

        // 解析并显示内容
        const content = parseSSE(chunkStr);
        if (content) {
          console.log('  >>> 内容: ' + content.replace(/\n/g, '\\n'));
        }

        // 统计总字节数
        totalBytes += chunk.length;
      });

      res.on('end', () => {
        const totalTime = Date.now() - totalStartTime;
        console.log('============================================================');
        console.log('流式传输完成！');
        console.log('============================================================');
        console.log('总耗时: ' + totalTime + 'ms');
        console.log('DNS 解析: ' + dnsTime + 'ms');
        console.log('TCP 连接: ' + tcpTime + 'ms');
        console.log('请求到响应: ' + requestToResponse + 'ms');
        console.log('数据块数量: ' + chunkCount);
        console.log('总字节数: ' + totalBytes);
        console.log('------------------------------------------------------------');
        console.log('各数据块时间线:');
        console.log('------------------------------------------------------------');

        // 计算间隔
        let prevTime = requestToResponse;
        chunkTimes.forEach((item) => {
          const interval = item.elapsed - prevTime;
          console.log('  #' + item.num.toString().padStart(2) + ': +' + interval.toString().padStart(5) + 'ms (' + item.bytes + ' bytes)');
          prevTime = item.elapsed;
        });

        console.log('------------------------------------------------------------');
        console.log('TTFT 分析:');
        const ttft = chunkTimes[0] ? (chunkTimes[0].elapsed - requestToResponse) : 0;
        console.log('  TTFT: ' + ttft + 'ms');
        console.log('  TTFT 占比: ' + ((ttft / totalTime) * 100).toFixed(1) + '%');
        console.log('============================================================');
      });
    });

    req.on('error', (error) => {
      console.error('请求错误: ' + error.message);
    });

    req.write(postData);
    req.end();
    console.log('[' + Date.now() + '] 请求已发送');
  });

  socket.on('error', (error) => {
    console.error('TCP 连接错误: ' + error.message);
  });
});

const net = require('net');
