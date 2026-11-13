import os
from dataclasses import dataclass
from pathlib import Path

NATIVE_ROOT = Path(__file__).resolve().parents[1] / 'native'
EXAMPLE_SOURCES = NATIVE_ROOT / 'sources_example'
GUO2_ADDED_SOURCES = NATIVE_ROOT / 'sources_custom'
OFFICIAL_SOURCES = NATIVE_ROOT.parent / 'guoapp-sources'
ENVIRONMENT_DEFINES = ['PROJECT_URL', 'TELEGRAM_GROUP_URL', 'TELEGRAM_CHANNEL_URL',
                       'ZHENGUOJIAN_DISPLAY_NAME', 'ZHENGUOJIAN_LATIN_NAME']


@dataclass(frozen=True)
class BuildVariant:
    all_sources: bool = False

    @property
    def slug(self):
        return 'zhenguojian' if self.all_sources else 'hongguojian'

    @property
    def bundle_identifier(self):
        if not self.all_sources:
            return 'com.guoapp.phoenix'
        return os.environ.get('ZHENGUOJIAN_PACKAGE_NAME', '').strip() or 'com.guoapp.all'

    @property
    def display_name(self):
        if not self.all_sources:
            return '红果鉴'
        return os.environ.get('ZHENGUOJIAN_DISPLAY_NAME', '').strip() or '真果鉴'

    @property
    def arguments(self):
        return ['--all-sources'] if self.all_sources else []

    @property
    def flutter_arguments(self):
        arguments = ['--dart-define=APP_FLAVOR=' + self.slug]
        for name in ENVIRONMENT_DEFINES:
            value = os.environ.get(name, '').strip()
            if value:
                arguments.append(f'--dart-define={name}={value}')
        return arguments

    @property
    def go_arguments(self):
        return ['-tags=' + self.slug, '-ldflags=' + self.linker_flags]

    @property
    def linker_flags(self):
        flags = '-s -w'
        update_server = os.environ.get('UPDATE_SERVER_URL', '').strip()
        if update_server:
            flags += ' -X duanjuapp/native/core.nativeUpdateServer=' + update_server
        project = os.environ.get('PROJECT_URL', '').strip()
        if project:
            flags += ' -X duanjuapp/native/core.nativeProjectURL=' + project
        return flags


def add_variant_argument(parser):
    parser.add_argument('--all-sources', action='store_true',
                        help='构建包含全部站源的全站源版；默认构建仅红果的红果鉴')


def add_sources_argument(parser):
    parser.add_argument('--sources', type=Path, action='append',
                        help='站源仓库路径，可多次指定；接入一次后会记住，之后可省略；从未接入时使用示例站源')


def source_module(directory):
    manifest = directory / 'go.mod'
    declared = manifest.read_text(encoding='utf-8').splitlines() if manifest.is_file() else []
    for line in declared:
        if line.startswith('module '):
            return line.removeprefix('module ').strip()
    raise SystemExit('站源仓库路径无效，找不到 go.mod 的 module 声明：' + str(directory))


def link_sources(paths):
    workspace = NATIVE_ROOT / 'go.work'
    anchor = NATIVE_ROOT / 'bridge' / 'sources_gen.go'
    require_official = os.environ.get('GUO_REQUIRE_OFFICIAL_SOURCES', '').strip() == '1'
    official = Path(
        os.environ.get('GUO_OFFICIAL_SOURCES', '').strip() or OFFICIAL_SOURCES
    ).expanduser().resolve()

    if not paths:
        if official.joinpath('go.mod').is_file():
            paths = [official, GUO2_ADDED_SOURCES]
        elif require_official:
            raise SystemExit(
                'Missing original 110 source module (guoapp/sources); '
                'refusing to create a full-sources APK with only 13 added '
                'channels. Configure authorized access to the upstream '
                'sources and place it at guoapp-sources.'
            )
        elif workspace.is_file() and anchor.is_file():
            return
        elif GUO2_ADDED_SOURCES.is_dir():
            print(
                'WARNING: only the 13 additional guo2 sources are present. '
                'Official 110 channels are NOT included; this is a '
                'partial-source development build.',
                flush=True,
            )
            paths = [GUO2_ADDED_SOURCES]
        else:
            paths = [EXAMPLE_SOURCES]

    if require_official:
        names = [source_module(path.expanduser().resolve()) for path in paths]
        if 'guoapp/sources' not in names:
            raise SystemExit(
                'Full APK blocked: official upstream source module '
                '(go.mod: module guoapp/sources) must be included.'
            )
        if 'duanjuapp/native/sources_custom' not in names:
            raise SystemExit(
                'Full APK blocked: guo2 extra source module '
                '(duanjuapp/native/sources_custom) must also be included.'
            )
    directories = [path.expanduser().resolve() for path in paths]
    modules = [source_module(directory) for directory in directories]
    uses = ''.join(f'\t{directory.as_posix()}\n' for directory in directories)
    workspace.write_text(f'go 1.24.1\n\nuse (\n\t.\n{uses})\n', encoding='utf-8')
    imports = ''.join(f'\t_ "{module}"\n' for module in modules)
    anchor.write_text(f'package main\n\nimport (\n{imports})\n', encoding='utf-8')
