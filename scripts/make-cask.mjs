// Generates the Homebrew cask for a release: one cask for macOS (Apple silicon, the app) and Linux (x64, the binary),
// so `brew install mormitech/tap/mcdiskfree` works on both. (A formula of the same name would win on macOS and fail.)
//
//   node scripts/make-cask.mjs [version] > ../homebrew-tap/Casks/mcdiskfree.rb        # release assets on GitHub
//   node scripts/make-cask.mjs --local dist > /tmp/mcdiskfree.rb                       # local test: file:// files in dist/
//
// The version defaults to the one in wails.json. The checksums are computed from the published files, not typed by hand.
import { createHash } from 'node:crypto';
import { existsSync, readFileSync } from 'node:fs';
import { join, resolve } from 'node:path';
import { pathToFileURL } from 'node:url';

const root = resolve(import.meta.dirname, '..');
const args = process.argv.slice(2);
const localAt = args.indexOf('--local');
const localDir = localAt >= 0 ? resolve(args[localAt + 1]) : null;
const version = args.find((a, i) => !a.startsWith('--') && i !== localAt + 1)
  ?? JSON.parse(readFileSync(join(root, 'wails.json'), 'utf8')).info.productVersion;

const repo = 'mormitech/mcdiskfree';
const files = {
  macos: `mcdiskfree-${version}-macos-arm64.zip`,
  linux: `mcdiskfree-${version}-linux-x64.tar.gz`,
};
const urlOf = (file) => (localDir ? pathToFileURL(join(localDir, file)).href : `https://github.com/${repo}/releases/download/v${version}/${file}`);

async function sha256(location) {
  let buf;
  if (location.startsWith('file:')) {
    if (!existsSync(new URL(location))) throw new Error(`${location} does not exist`);
    buf = readFileSync(new URL(location));
  } else {
    const res = await fetch(location);
    if (!res.ok) throw new Error(`${location}: HTTP ${res.status} (is the release published?)`);
    buf = Buffer.from(await res.arrayBuffer());
  }
  return createHash('sha256').update(buf).digest('hex');
}

// In a local test the url is a file:// one; a published cask always uses the #{version} form.
const urlLine = (file) => (localDir ? urlOf(file) : `https://github.com/${repo}/releases/download/v#{version}/${file.replace(version, '#{version}')}`);

process.stdout.write(`cask "mcdiskfree" do
  version "${version}"

  name "mcDiskFree"
  desc "Disk space and disk usage analyzer with an interactive sunburst chart"
  homepage "https://github.com/${repo}"

  livecheck do
    url :url
    strategy :github_latest
  end

  on_macos do
    sha256 "${await sha256(urlOf(files.macos))}"

    url "${urlLine(files.macos)}"

    depends_on arch: :arm64
    depends_on macos: :monterey

    app "mcdiskfree.app"
    # also a command, so mcdiskfree (and mcdiskfree --mock macos) works in a terminal
    binary "#{appdir}/mcdiskfree.app/Contents/MacOS/mcdiskfree"

    # The app is not notarized yet, so Gatekeeper would block its first launch; clear the quarantine flag it got
    # when it was downloaded. (A third-party tap may do this; the main homebrew-cask repository does not allow it.)
    postflight_steps do
      run "/usr/bin/xattr",
          args:         ["-dr", "com.apple.quarantine", "{{appdir}}/mcdiskfree.app"],
          must_succeed: false
    end

    zap trash: "~/Library/Caches/mcdiskfree"
  end

  on_linux do
    sha256 "${await sha256(urlOf(files.linux))}"

    url "${urlLine(files.linux)}"

    depends_on arch: :x86_64

    binary "mcdiskfree"

    zap trash: "~/.cache/mcdiskfree"

    caveats <<~EOS
      mcdiskfree needs GTK 3 and WebKitGTK 4.1 from your system, for example on Debian and Ubuntu:
        sudo apt install libgtk-3-0 libwebkit2gtk-4.1-0
    EOS
  end
end
`);
