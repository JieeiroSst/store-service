import io
import os
import subprocess
import sys
import tempfile

from PIL import Image

HERE = os.path.dirname(os.path.abspath(__file__))
APP = os.path.abspath(os.path.join(HERE, '..', '..'))
REPO = os.path.abspath(os.path.join(APP, '..'))
BRAND = os.path.join(APP, 'assets', 'brand')
WEB_BOOK = os.path.join(REPO, 'keikibook', 'public')


def svg_source(name):
    with open(os.path.join(BRAND, name)) as f:
        return f.read()


def framed(inner_svg, scale, background=None):
    body = inner_svg.split('>', 1)[1].rsplit('</svg>', 1)[0]
    offset = 512 * (1 - scale)
    bg = background.split('>', 1)[1].rsplit('</svg>', 1)[0] if background else ''
    return (
        '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1024 1024">'
        f'{bg}<g transform="translate({offset} {offset}) scale({scale})">{body}</g></svg>'
    )


def render(svg_text, size):
    with tempfile.NamedTemporaryFile('w', suffix='.svg', delete=False) as src:
        src.write(svg_text)
    out = src.name + '.png'
    subprocess.run(['node', os.path.join(HERE, 'render.mjs'), src.name, out, str(size)], check=True)
    img = Image.open(out).convert('RGBA')
    img.load()
    os.remove(src.name)
    os.remove(out)
    return img


def save(img, path, opaque=False):
    os.makedirs(os.path.dirname(path), exist_ok=True)
    if opaque:
        flat = Image.new('RGB', img.size, (67, 56, 202))
        flat.paste(img, mask=img.split()[3])
        img = flat
    img.save(path, optimize=True)
    print('wrote', os.path.relpath(path, REPO), img.size)


def main():
    icon = svg_source('keikibook_icon.svg')
    square = svg_source('keikibook_icon_square.svg')
    mark = svg_source('keikibook_mark.svg')
    mono = svg_source('keikibook_monochrome.svg')
    background = svg_source('keikibook_background.svg')
    cache = {}

    def r(key, svg_text, size):
        k = (key, size)
        if k not in cache:
            cache[k] = render(svg_text, size)
        return cache[k]

    res = os.path.join(APP, 'android', 'app', 'src', 'main', 'res')
    for dpi, px in {'mdpi': 48, 'hdpi': 72, 'xhdpi': 96, 'xxhdpi': 144, 'xxxhdpi': 192}.items():
        save(r('icon', icon, px), os.path.join(res, f'mipmap-{dpi}', 'ic_launcher.png'))
        fg = px * 108 // 48
        save(r('fg', framed(mark, 0.76), fg), os.path.join(res, f'drawable-{dpi}', 'ic_launcher_foreground.png'))
        save(r('mono', framed(mono, 0.76), fg), os.path.join(res, f'drawable-{dpi}', 'ic_launcher_monochrome.png'))
        save(r('splash', framed(mark, 1.0), px * 2), os.path.join(res, f'drawable-{dpi}', 'splash_mark.png'))

    ios = os.path.join(APP, 'ios', 'Runner', 'Assets.xcassets', 'AppIcon.appiconset')
    for name, px in {
        'Icon-App-20x20@1x.png': 20, 'Icon-App-20x20@2x.png': 40, 'Icon-App-20x20@3x.png': 60,
        'Icon-App-29x29@1x.png': 29, 'Icon-App-29x29@2x.png': 58, 'Icon-App-29x29@3x.png': 87,
        'Icon-App-40x40@1x.png': 40, 'Icon-App-40x40@2x.png': 80, 'Icon-App-40x40@3x.png': 120,
        'Icon-App-60x60@2x.png': 120, 'Icon-App-60x60@3x.png': 180,
        'Icon-App-76x76@1x.png': 76, 'Icon-App-76x76@2x.png': 152, 'Icon-App-83.5x83.5@2x.png': 167,
        'Icon-App-1024x1024@1x.png': 1024,
    }.items():
        save(r('square', square, px), os.path.join(ios, name), opaque=True)
    launch = os.path.join(APP, 'ios', 'Runner', 'Assets.xcassets', 'LaunchImage.imageset')
    for name, px in {'LaunchImage.png': 120, 'LaunchImage@2x.png': 240, 'LaunchImage@3x.png': 360}.items():
        save(r('icon', icon, px), os.path.join(launch, name))

    mac = os.path.join(APP, 'macos', 'Runner', 'Assets.xcassets', 'AppIcon.appiconset')
    for px in (16, 32, 64, 128, 256, 512, 1024):
        save(r('mac', framed(icon, 0.82), px), os.path.join(mac, f'app_icon_{px}.png'))

    web = os.path.join(APP, 'web')
    save(r('icon', icon, 32), os.path.join(web, 'favicon.png'))
    for px in (192, 512):
        save(r('icon', icon, px), os.path.join(web, 'icons', f'Icon-{px}.png'))
        save(r('maskable', framed(mark, 0.9, background), px), os.path.join(web, 'icons', f'Icon-maskable-{px}.png'))

    if os.path.isdir(WEB_BOOK):
        for px in (192, 512):
            save(r('icon', icon, px), os.path.join(WEB_BOOK, f'logo{px}.png'))
        save(r('icon', icon, 180), os.path.join(WEB_BOOK, 'apple-touch-icon.png'))
        ico = r('icon', icon, 256)
        ico.save(os.path.join(WEB_BOOK, 'favicon.ico'), sizes=[(16, 16), (32, 32), (48, 48), (64, 64)])
        print('wrote', os.path.relpath(os.path.join(WEB_BOOK, 'favicon.ico'), REPO))
        with open(os.path.join(WEB_BOOK, 'logo.svg'), 'w') as f:
            f.write(icon)


if __name__ == '__main__':
    sys.exit(main())
