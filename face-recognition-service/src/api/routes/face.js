const Router = require('@koa/router');
const multer = require('@koa/multer');

const router = new Router();
const upload = multer();

// Latest face / color positions reported by the tracking.js page
let latestPositions = { faces: [], colors: [], frame: null, updatedAt: null };

const toRect = ({ x, y, width, height }) => ({
  x: Number(x), y: Number(y), width: Number(width), height: Number(height),
});

router.post('/face/position', ctx => {
  const { faces = [], colors = [], frame } = ctx.request.body || {};
  if (!Array.isArray(faces) || !Array.isArray(colors)) {
    ctx.status = 400;
    ctx.body = { error: 'faces and colors must be arrays' };
    return;
  }

  latestPositions = {
    faces: faces.map(toRect),
    colors: colors.map(c => ({ color: String(c.color), ...toRect(c) })),
    frame,
    updatedAt: new Date().toISOString(),
  };
  console.log('position', JSON.stringify(latestPositions));
  ctx.body = { ok: true };
});

router.get('/face/position', ctx => {
  ctx.body = latestPositions;
});

router.post(
  '/upload-multiple-files',
  upload.fields([
    {
      name: 'avatar',
      maxCount: 1
    },
    {
      name: 'boop',
      maxCount: 2
    }
  ]),
  ctx => {
    console.log('ctx.request.files', ctx.request.files);
    console.log('ctx.files', ctx.files);
    console.log('ctx.request.body', ctx.request.body);
    ctx.body = 'done';
  }
);

router.post(
  '/upload-single-file',
  upload.single('avatar'),
  ctx => {
    console.log('ctx.request.file', ctx.request.file);
    console.log('ctx.file', ctx.file);
    console.log('ctx.request.body', ctx.request.body);
    ctx.body = 'done';
  }
);

module.exports = router;