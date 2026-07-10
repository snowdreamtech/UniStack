import urllib.request
import zipfile
import os
import glob

# Try to download the workflow logs zip
url = "https://github.com/snowdreamtech/UniStack/actions/runs/29070537998/logs"
# Unfortunately, GitHub API requires authentication to download logs even for public repos using the /logs endpoint.
# But we can try!
