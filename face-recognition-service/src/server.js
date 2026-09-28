const fs = require('fs');
const path = require('path');
const Koa = require('koa');
const bodyParser = require('koa-bodyparser');
const faceRouter = require('./api/routes/face');
const app = new Koa();

const PUBLIC_DIR = path.join(__dirname, '..', 'public');

app.use(async (ctx, next) => {
  try {
    await next();
  } catch (err) {
    console.error(err);
    ctx.status = err.status || 500;
    ctx.body = { error: err.message };
  }
});

app.use(bodyParser());

faceRouter.prefix('/api');
app.use(faceRouter.routes()).use(faceRouter.allowedMethods());

app.use(ctx => {
  if (ctx.path === '/' || ctx.path === '/face-tracking') {
    ctx.type = 'html';
    ctx.body = fs.createReadStream(path.join(PUBLIC_DIR, 'index.html'));
    return;
  }
  ctx.body = 'Hello World';
});

app.listen(1234);
