import collections
import json
import os
from pathlib import Path

path = Path(os.environ['RUNNER_TEMP']) / (os.environ['SERVICE'] + '-tests.jsonl')
if not path.exists():
    raise SystemExit('Test report was not produced; inspect the failed prior step.')
counts = collections.Counter()
for line in path.read_text().splitlines():
    item = json.loads(line)
    if item.get('Test') and item.get('Action') in ('pass', 'fail', 'skip'):
        counts[item['Action']] += 1
    if item.get('Action') == 'fail' and item.get('Output'):
        print(item['Output'].rstrip())
summary = f"{os.environ['SERVICE']}: {counts['pass']} tests passed, {counts['fail']} failed, {counts['skip']} skipped"
print(summary)
with open(os.environ['GITHUB_STEP_SUMMARY'], 'a') as output:
    output.write(summary + '\n')
if counts['fail'] or counts['skip']:
    raise SystemExit('Release/integration tests must pass without skipped DB coverage.')
