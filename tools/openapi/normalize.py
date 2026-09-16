import json
import pathlib
import sys


spec_path = pathlib.Path(sys.argv[1])
spec = json.loads(spec_path.read_text())
for methods in spec['paths'].values():
    for operation in methods.values():
        content = operation.get('requestBody', {}).get('content', {})
        form = content.get('application/x-www-form-urlencoded', {}).get('schema', {})
        if 'multipart/form-data' not in content or form.get('type') != 'file':
            continue
        field = form['title']
        content['multipart/form-data']['schema'] = {
            'type': 'object',
            'required': [field] if operation['requestBody'].get('required') else [],
            'properties': {field: {'type': 'string', 'format': 'binary'}},
        }
        del content['application/x-www-form-urlencoded']

spec_path.write_text(json.dumps(spec, indent=4) + '\n')
