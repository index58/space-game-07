const http = require("http");
const fs = require("fs");
const path = require("path");

const port = Number(process.env.TRADE_MOCKUP_PORT || 62936);
const host = process.env.TRADE_MOCKUP_HOST || "127.0.0.1";
const root = __dirname;

// Отдаёт единственный HTML-макет без запуска игрового клиента.
const server = http.createServer((request, response) => {
  const requestedPath = request.url === "/" ? "/index.html" : request.url;
  const filePath = path.join(root, path.basename(requestedPath));

  if (!fs.existsSync(filePath)) {
    response.writeHead(404, { "content-type": "text/plain; charset=utf-8" });
    response.end("Not found");
    return;
  }

  response.writeHead(200, { "content-type": "text/html; charset=utf-8" });
  response.end(fs.readFileSync(filePath));
});

server.listen(port, host, () => {
  console.log(`http://${host}:${port}`);
});
