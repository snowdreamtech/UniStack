import re
with open('cmd/0.root.go', 'r') as f:
    content = f.read()

if '"context"' not in content:
    content = re.sub(
        r'import \(',
        'import (\n\t"context"\n\t"time"',
        content
    )
    with open('cmd/0.root.go', 'w') as f:
        f.write(content)
