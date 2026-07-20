import os
import glob
import re
import subprocess

def get_old_content(os_name):
    try:
        content = subprocess.check_output(['git', 'show', f'HEAD~1:ansible/roles/apps/foundation/vars/{os_name}.yml']).decode('utf-8')
        return content
    except:
        return ""

os_files = [os.path.basename(f) for f in glob.glob('ansible/roles/apps/foundation/vars/*.yml')]

for f in os_files:
    os_name = f.replace('.yml', '')
    old_content = get_old_content(os_name)
    if not old_content:
        continue
    
    lines = old_content.split('\n')
    extracted_lines = ["---"]
    capturing = False
    for line in lines:
        if line.startswith('init_openssh_packages:') or line.startswith('init_sshd_service:') or line.startswith('init_sshd_use_pam:'):
            extracted_lines.append(line.replace('init_', 'openssh_').replace('openssh_openssh_', 'openssh_'))
            capturing = True
        elif capturing and line.startswith('  -'):
            extracted_lines.append(line)
        else:
            capturing = False
            
    if len(extracted_lines) > 1:
        with open(f'ansible/roles/apps/openssh/vars/{f}', 'w') as out_f:
            out_f.write('\n'.join(extracted_lines) + '\n')
