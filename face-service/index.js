const Koa = require('koa');
const render = require('koa-ejs');
const path = require('path');
const static = require('koa-static-router');
const logger = require('koa-logger')

const {CameraRouter} =require("./internal/delivery/http");

const app = new Koa();

app.use(logger())
app.use(static([
    {
        dir: 'public',
        router: '/public/'
    }, {
        dir: 'static',
        router: '/static/'
    }
]))

render(app, {
    root: path.join(__dirname, 'view'),
    layout: 'template',
    viewExt: 'html',
    cache: false,
    debug: true
});

app.use(CameraRouter.routes());

const PORT = process.env.PORT || 1234;
app.listen(PORT, () => console.log(`running on port ${PORT}`));
