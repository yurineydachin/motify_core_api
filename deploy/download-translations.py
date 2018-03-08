import subprocess
import sys
from StringIO import StringIO
from posixpath import join as urljoin
import json
import os
import re
import zipfile
import urllib2
import time
import glob


def download_translations(src_path=sys.argv[1], wti_token=sys.argv[2]):
    """
    Task downloads translations from webtranslateit and converts them to json
    """
    sys.exit(0)

    api_url = "https://webtranslateit.com/api"
    project_url = urljoin(api_url, "projects", wti_token+".json")
    zip_url = urljoin(api_url, "projects", wti_token, "zip_file")

    locale_aliases = {
        "en_en_VN": "en_VN",
        "en_en_TH": "en_TH",
        "en_en_ID": "en_ID",
        "ms":       "ms_MY",
    }
    php_re = re.compile(r'\$lang\["(?P<key>.*?)"\] = "(?P<value>.*?)";\r?\n', re.S | re.I)
    translations_path = os.path.join(src_path, "translations")

    tries_counter = 3
    while tries_counter > 0:
        try:
            print("Try to load traslations (#{0})".format((4-tries_counter)))
            project = json.load(urllib2.urlopen(project_url))['project']
            zip_file = zipfile.ZipFile(StringIO(urllib2.urlopen(zip_url).read()))

            if not('project' in locals() or 'project' in globals()) or not('zip_file' in locals() or 'zip_file' in globals()):
                print("Translation loading is failed")
                exist_translations = filter(os.path.isfile, glob.glob("{0}/*.json".format(translations_path)))
                if len(exist_translations) > 0:
                    print("Use exist old translation files.")
                    sys.exit(0)
                else:
                    sys.exit(1)

            # cleaning previous translations
            print("Remove old translations... \n$rm -f {0}/*.json".format(translations_path))
            subprocess.call("rm -f {0}/*.json".format(translations_path), shell=True)

            for file_info in project['project_files']:
                if not file_info['name'].endswith('/alice.php'):
                    continue
                locale = locale_aliases.get(file_info['locale_code'], file_info['locale_code'])
                data = {
                    "hash": file_info['hash_file'],
                    "locale": locale,
                    "phrases": dict(),
                }
                for key, value in php_re.findall(zip_file.read(file_info['name'])):
                    if key:
                        key = key.replace(r'\"', '"')
                        data['phrases'][key] = value.replace(r'\"', '"') or key

                translation_path = os.path.join(translations_path, locale+".json")
                with open(translation_path, 'w') as f:
                    print("write translations into {0}".format(translation_path))
                    json.dump(data, f, ensure_ascii=False)

            print("Translations are downloaded and written")
            break

        except Exception as error:
            print('Failed with error: ' + repr(error))
            tries_counter -= 1
            time.sleep(2)

    # final check if translation files are exist
    exist_translations = filter(os.path.isfile, glob.glob("{0}/*.json".format(translations_path)))
    if len(exist_translations) == 0:
        print('Translations are not downloaded.')
        sys.exit(1)


if __name__ == '__main__':
    download_translations()
