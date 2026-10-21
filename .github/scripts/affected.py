"""Select existing service checks; release candidates always check everything."""
import json
import os
from pathlib import Path
import subprocess

BACKENDS = ['core-backend', 'global-backend']
FRONTENDS = ['admin-web', 'core-web', 'organizer-web', 'public-web']


def command(args):
    return subprocess.check_output(args, text=True).strip()


def main():
    event = json.loads(Path(os.environ['GITHUB_EVENT_PATH']).read_text())
    full = os.environ.get('FULL_CHECKS') == 'true'
    skip = False
    kind = os.environ.get('GITHUB_EVENT_NAME')
    if kind == 'push' and not full and os.environ.get('GITHUB_REF_TYPE') == 'branch':
        repo = os.environ['GITHUB_REPOSITORY']
        owner = repo.split('/')[0]
        head = owner + ':' + os.environ['GITHUB_REF_NAME']
        # PR CI owns these checks. A push run must not duplicate its builds.
        count = command(['gh', 'api', '--method', 'GET', f'repos/{repo}/pulls', '-f', f'head={head}', '-f', 'state=open', '--jq', 'length'])
        skip = int(count) > 0
    if 'pull_request' in event:
        base = event['pull_request']['base']['sha']
        head = event['pull_request']['head']['sha']
    else:
        base, head = event.get('before', ''), os.environ['GITHUB_SHA']
    if not base or set(base) == {'0'}:
        full = True
    paths = [] if full or skip else command(['git', 'diff', '--name-only', base + '...' + head]).splitlines()
    shared = any(p.startswith(('.github/', 'deploy/migrations/')) or p in ('AGENTS.md', '.dockerignore') for p in paths)
    front_shared = any(p.startswith('packages/') or p in ('package.json', 'package-lock.json') for p in paths)
    backends = [s for s in BACKENDS if full or shared or any(p.startswith(f'services/{s}/') for p in paths)]
    frontends = [s for s in FRONTENDS if full or shared or front_shared or any(p.startswith(f'apps/{s}/') for p in paths)]
    operations = full or shared or any(p.startswith('deploy/') for p in paths)
    result = {'skip': str(skip).lower(), 'backends': json.dumps(backends if not skip else []), 'frontends': json.dumps(frontends if not skip else []), 'operations': str(operations and not skip).lower()}
    with open(os.environ['GITHUB_OUTPUT'], 'a') as output:
        for key, value in result.items():
            output.write(f'{key}={value}\n')
    print('Selected checks:', result)


if __name__ == '__main__':
    main()
