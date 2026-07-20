import os
import re
import sys

def process_file(filepath):
    with open(filepath, 'r') as f:
        content = f.read()

    # Match: success_msg: "something validated ({{ var_name }})"
    # We want to replace {{ var_name }} with {{ var_name | default('UNDEFINED') }}
    # But only if it doesn't already have a pipe |
    
    def repl(m):
        full_match = m.group(0)
        var_name = m.group(1).strip()
        
        if '|' in var_name:
            return full_match
            
        return full_match.replace(f"{{{{ {var_name} }}}}", f"{{{{ {var_name} | default('UNDEFINED') }}}}")

    new_content = re.sub(r'success_msg:\s*"[^"]*validated \(\{\{\s*([^}]+)\s*\}\}\)"', repl, content)
    
    if new_content != content:
        with open(filepath, 'w') as f:
            f.write(new_content)
        print(f"Fixed {filepath}")

for root, _, files in os.walk('ansible/roles'):
    for file in files:
        if file.endswith('.yml'):
            process_file(os.path.join(root, file))
