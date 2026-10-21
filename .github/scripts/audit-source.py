"""Reject forbidden tracked artifacts without printing sensitive matches."""
from pathlib import Path
import re
import subprocess


def main():
    paths = subprocess.check_output(['git', 'ls-files', '-z']).decode().split('\0')
    findings = []
    key_pattern = re.compile(rb'-----BEGIN (?:[A-Z0-9]+ )*PRIVATE KEY-----')
    token_pattern = re.compile(rb'\b(?:gh[pousr]_[A-Za-z0-9]{30,}|github_pat_[A-Za-z0-9_]{50,}|AKIA[A-Z0-9]{16})\b')
    forbidden_dirs = {'node_modules', '.next', 'logs', 'coverage', '__pycache__'}
    for name in filter(None, paths):
        path = Path(name)
        if (any(p in forbidden_dirs for p in path.parts)
                or path.name.startswith('.env') and path.name != '.env.example'
                or path.suffix.lower() in {'.pem', '.key', '.pfx', '.p12', '.dump', '.zip', '.tar', '.gz', '.exe', '.tsbuildinfo'}):
            findings.append((name, 'forbidden tracked file'))
            continue
        if path.is_file():
            content = path.read_bytes()
            if key_pattern.search(content) or token_pattern.search(content):
                findings.append((name, 'credential-shaped content'))
    if findings:
        for name, reason in findings:
            print(f'{name}: {reason}; matched value withheld')
        raise SystemExit(1)
    print('Tracked source audit passed; generated SQLc query contracts are intentionally versioned.')


if __name__ == '__main__':
    main()
