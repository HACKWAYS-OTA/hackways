// Build-time asset utility. Uses the same React presentation component as live OG.
const fs = require('node:fs');
const path = require('node:path');
const Module = require('node:module');
const ts = require('typescript');
const React = require('react');
const { ImageResponse } = require('next/og');
const root = path.resolve(__dirname, '../..');
function sourceModule(relative) {
  const file = path.join(root, relative), mod = new Module(file, module);
  mod.filename = file; mod.paths = Module._nodeModulePaths(path.dirname(file));
  mod._compile(ts.transpileModule(fs.readFileSync(file, 'utf8'), { compilerOptions: { module: ts.ModuleKind.CommonJS, jsx: ts.JsxEmit.ReactJSX, target: ts.ScriptTarget.ES2020 } }).outputText, file);
  return mod.exports;
}
async function main() {
  const { SocialImage, socialSizes } = sourceModule('packages/ui/src/SocialImage.tsx');
  const { socialPages } = sourceModule('packages/ui/src/SocialPages.ts');
  const { faqTopics } = sourceModule('apps/public-web/lib/help-topics.ts');
  const cards = { ...socialPages };
  for (const topic of faqTopics) cards[`help-${topic.id}`] = { title: topic.label, description: topic.intro, label: 'Help & FAQs' };
  const fonts = [['SocialDisplay', 'social-display.ttf'], ['SocialBody', 'social-body.ttf']].map(([name, file]) => ({ name, data: new Uint8Array(fs.readFileSync(path.join(root, 'apps/core-web/public/fonts', file))).buffer, weight: 400, style: 'normal' }));
  const output = path.join(root, 'apps/public-web/public/social'); fs.mkdirSync(output, { recursive: true });
  for (const [key, content] of Object.entries(cards)) {
    for (const format of ['wide', 'square']) {
      const response = new ImageResponse(React.createElement(SocialImage, { content, format }), { ...socialSizes[format], fonts });
      const bytes = Buffer.from(await response.arrayBuffer());
      fs.writeFileSync(path.join(output, `${key}-${format}.png`), bytes);
    }
    process.stdout.write(`${key}: wide + square\n`);
  }
  fs.writeFileSync(path.join(output, 'PROVENANCE.md'), 'These static PNG previews are rendered from packages/ui/src/SocialImage.tsx and SocialPages.ts by deploy/scripts/render-social-images.cjs. They use the existing Hackways brand fonts. The ambassador ASCII motif is code-rendered from the user-approved campaign direction. No application records or private data are fabricated. Regenerate with node deploy/scripts/render-social-images.cjs.\n');
}
main().catch(error => { console.error('Static social image rendering failed:', error.message); process.exitCode = 1; });
