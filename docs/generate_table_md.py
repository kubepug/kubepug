'''
Python script to generate the api status page
'''
import os
import json
from datetime import datetime
from mdutils.mdutils import MdUtils

HIDE_NAVIGATION='''---
hide:
  - navigation
  - toc
---
'''

def get_replacements(replacement_item: dict) -> str:
    """function to get a replacement_item and generate a table cell entry"""
    if len(replacement_item) == 0:
        return ''

    replacement_content = (replacement_item['group'] + '/' +
        replacement_item['version'] + '/' +
        replacement_item['kind'])
    return replacement_content

with open('docs/data/data.json', encoding='utf-8') as apidatafile:
    data = json.load(apidatafile)

    mdFile = MdUtils(file_name='status')
    tablecontent = [
        "Group", 
        "Version", 
        "Kind", 
        "Deprecated", 
        "Deleted", 
        "Replacement"
    ]

    len_table: int = 1
    for items in data:
        group = items['group']
        version = items['version']
        kind = items['kind']
        if 'replacement' in items:
            replacement = get_replacements(items['replacement'])
        deprecated: str = (str(items['deprecated_version']['version_major']) + '.' +
                           str(items['deprecated_version']['version_minor']))

        deleted: str = (str(items['removed_version']['version_major']) + '.' +
                        str(items['removed_version']['version_minor']))

        tablecontent.extend([group, version, kind, deprecated, deleted, replacement])
        len_table: int = len_table + 1

    mdFile.write('Page generated at ' + datetime.now().strftime("%Y-%b-%d %H:%M:%S"))
    mdFile.new_line()
    mdFile.new_table(columns=6, rows=len_table, text=tablecontent, text_align='left')
    mdFile.new_line()
    content = HIDE_NAVIGATION + mdFile.file_data_text

    with open('docs/' + mdFile.file_name + '.md', 'w', encoding='utf-8') as md:
        md.write(content)
