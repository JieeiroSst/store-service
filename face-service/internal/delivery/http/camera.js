const Router = require('@koa/router');

const CameraRouter = new Router();

CameraRouter.get('/', async (ctx) => {
    const host = ctx.request.host;
    console.log(host);
    await ctx.render("face");
});

module.exports = CameraRouter;
