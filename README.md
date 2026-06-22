# A small CLI note taking tool with your favorite editor

> [!NOTE]
> これは、MIT ライセンスの [rhysd/notes-cli](https://github.com/rhysd/notes-cli) をもとに独自にメンテナンスしているフォークです。

notes-cliは、ターミナル上でお気に入りのエディタを使ってメモを取るための小さなCLIツールです。  
ターミナルからこのツールを介して、メモの管理（作成/閲覧/一覧表示）を行うことができます。
また、メモの紛失を防ぐために、オプションでGitを使ってメモを保存することも可能です。  

このツールは、`grep`（または [ag](https://github.com/ggreer/the_silver_searcher), [rg](https://github.com/BurntSushi/ripgrep)）や `rm`、[fzf](https://github.com/junegunn/fzf) や [peco](https://github.com/peco/peco) などのフィルタリングツール、そしてコマンドラインから起動できるエディタなど、他のコマンドと組み合わせて快適に使用できるように設計されています。

このツールは、ObsidianとHugoとの互換性を重視しており、作成したノートを静的サイトにビルドすることができます。


## 目次

* [インストール](#インストール)
* [基本的な使い方](#基本的な使い方)
* [使い方](#使い方)
* [FAQ](#faq)
* [ライセンス](#ライセンス)

## インストール

以下のように、ソースから直接ビルドしてインストールできます。Goのツールチェーン（バージョン1.16以降）が必要です。
NixOSの場合、flakeをつかってビルドすることもできますが、更新がむずかしく推奨されません。

```
$ go install github.com/asano69/notes-cli/cmd/notes

```

実際に使い始める前に、サンプルを使って試すことができます。

```sh
$ git clone https://github.com/asano69/notes-cli.git
$ cd notes-cli/
$ export NOTES_CLI_HOME="$(pwd)/example/notes-cli"
$ export NOTES_CLI_EDITOR=vim # お気に入りのエディタを設定
$ notes list --full
$ notes new test my-local-trial
$ git status # ホームディレクトリにどのようなファイルが作成されたか確認

```

アンインストールする場合：

```sh
$ rm -rf "$(notes config home)" # すべてのメモを削除
$ rm "$(which notes)" # 実行ファイルを削除

```

## 基本的な使い方

`notes` は、Markdown形式のメモを管理するためのサブコマンドをいくつか提供しています。

* `notes new <category> <filename> [<tags>]` で新しいメモを作成します。すべてのメモは必ず1つのカテゴリに属する必要があり、タグは0個以上指定できます。
* `notes ls -e` を実行すると、既存のメモをお気に入りのエディタで開くことができます。この場合、`$NOTES_CLI_EDITOR`（設定されていない場合は代替として `EDITOR`）が設定されている必要があります。
* `notes ls -o` を実行すると、既存のメモをターミナル上で確認できます（`-o` は各メモの情報を1行で表示することを意味します）。

notes-cliのホームディレクトリ下のディレクトリ構造は、以下のようになります。

```
<HOME>
├── category1
│   ├── nested-category
│   │   └── note3.md
│   ├── note1.md
│   └── note2.md
├── category2
│   ├── note4.md
│   └── note5.md
└── category3
    └── note6.md

```

ここで、`<HOME>` はデフォルトで [XDG Data ディレクトリ](https://wiki.archlinux.org/index.php/XDG_Base_Directory)（macOSの場合は `~/.local/share/notes-cli`）となり、環境変数 `$NOTES_CLI_HOME` で指定することもできます。

より実践的なホームディレクトリの例は、[example ディレクトリ](./example/notes-cli) で確認できます。

## 使い方

このセクションでは、各操作の詳細な使い方を説明します。

* [新しいメモを作成する](#新しいメモを作成する)
* [作成したメモを柔軟に開く](#作成したメモを柔軟に開く)
* [作成したメモをリストとして確認する](#作成したメモをリストとして確認する)
* [メモのテンプレート](#メモのテンプレート)
* [Gitリポジトリにメモを保存する](#gitリポジトリにメモを保存する)
* [環境変数で動作を設定する](#環境変数で動作を設定する)
* [新しいサブコマンドを追加して notes コマンドを拡張する](#新しいサブコマンドを追加して-notes-コマンドを拡張する)
* [シェルでの補完](#シェルでの補完)
* [man マニュアルの設定](#man-マニュアルの設定)
* [ツール自体のアップデート](#ツール自体のアップデート)
* [Goプログラムから使用する](#goプログラムから使用する)

### 新しいメモを作成する

たとえば、次のように実行します。

```
$ notes new blog how-to-handle-files golang,file

```

これにより、`<HOME>/notes-cli/blog/how-to-handle-files.md` にメモファイルが作成されます。ホームディレクトリは自動的に作成されます。

カテゴリは `blog` です。すべてのメモは必ず1つのカテゴリに属している必要があります。カテゴリは `/` を使って階層化（ネスト）できます。たとえば、Blog A と Blog B のように複数のブログを運営している場合、`blog/A`、`blog/B` のようなカテゴリでブログ記事を分類すると便利です。

タグは `golang` と `file` です。タグはメモを整理し、検索しやすくするためのラベルです。タグは省略可能です。

隠しファイルや隠しディレクトリが作成されないよう、カテゴリ名とファイル名の先頭に `.` を使用することはできません。

環境変数 `$NOTES_CLI_EDITOR` にお気に入りのエディタを設定している場合、新しく作成されたメモファイルがそのエディタで開きます。そのままシームレスにファイルを編集できます。（設定されていない場合は `$EDITOR` も参照されます。）

```markdown
---
title: "Hello World!"
summary: "Hugo is a static site generator written in Go. It converts Markdown 
  files into HTML with remarkable speed."
tags: [golang, hugo]
categories: [wiki]
draft:
date: 2025-06-01T09:45:01+09:00
lastmod: 2026-06-19T17:48:41+09:00
---
```

`category: ...`、`tags: ...`、`created: ...` の行、およびタイトル（`# how-to-handle-files`）は削除しないでください。これらは `notes` コマンドによって使用されます（内容を変更する分には問題ありません）。デフォルトのタイトルはファイル名になります。以下のように、メモのタイトルや本文を自由に編集できます。

```markdown
---
title: "Hello World!"
summary: "Hugo is a static site generator written in Go. It converts Markdown 
  files into HTML with remarkable speed."
tags: [golang, hugo]
categories: [wiki]
draft:
date: 2025-06-01T09:45:01+09:00
lastmod: 2026-06-19T17:48:41+09:00
---

ドキュメントを読んでください。
GoDocがすべてを解説しています。
```

すべてのメモは、そのメモのカテゴリディレクトリ配下に配置されます。メモのカテゴリを変更する場合は、手動でディレクトリ構造を調整（メモファイルを新しいカテゴリのディレクトリへ移動）する必要があります。

詳細については、`notes new --help` を確認してください。

### 作成したメモを柔軟に開く

作成したメモをいくつか開く方法について説明します。

以下のコマンドで、メモのパス一覧を表示できます。

```
$ notes list # または `notes ls`

```

たとえば、現時点でメモが1つしかない場合は、以下のように1つのパスが表示されます。

```
/Users/me/.local/share/notes-cli/blog/how-to-handle-files.md

```

なお、`/Users/<ユーザー名>/.local/share` は macOS や Linux におけるデフォルトの XDG data ディレクトリです。環境変数 `$NOTES_CLI_HOME` を設定することで、この場所を変更できます。

一覧表示されたメモをエディタで開くには、`--edit`（または `-e`）を使用するのが最も手っ取り早い方法です。

```
$ notes list --edit
$ notes ls -e

```

複数のメモがある場合、メモは1行ずつ出力されます。そのため、`grep`、`head`、`peco`、`fzf` などを使ってリストをフィルタリングすることで、特定のメモを簡単に取り出すことができます。

```
$ notes ls | grep -l file | xargs -o vim

```

または、以下のような方法も機能します。

```
vim $(notes ls | xargs grep file)

```

また、`grep`、`rg`、`ag` などを使えば、メモの検索も簡単です。

```
$ notes ls | xargs ag documentation

```

検索して、それをそのまま Vim で開きたい場合も簡単です。

```
$ notes ls | xargs ag -l documentation | xargs -o vim

```

`notes ls` は `--sort` オプションを受け付け、リストの順序を変更できます。デフォルトの順序は、メモの作成日時の降順（新しい順）です。メモの更新日時（modified）で並び替えると、最新の更新ファイルが1行目に出力されるため、以下のようにして最後に編集したメモを即座に開くことができます。

```
$ note ls --sort modified | head -1 | xargs -o vim

```

詳細については、`notes list --help` を確認してください。

### 作成したメモをリストとして確認する

`notes list` は、メモの概要をターミナルに表示することもできます。

`--full`（または `-f`）オプションを使用すると、メモの全情報をターミナルに表示できます。

```
$ notes list --full

```

出力例：

```
/Users/me/.local/share/notes-cli/blog/how-to-handle-files.md
- Category: blog
- Tags: golang, file
- Created: 2018-10-28T07:19:27+09:00

How to handle files in Go
=========================

ドキュメントを読んでください。
GoDocがすべてを解説しています。


```

このオプションでは、以下の情報が色付きで表示されます。

* メモファイルへのフルパス
* メタデータ（`Category`、`Tags`、`Created`）
* メモのタイトル
* メモの本文（最大10行まで）

出力が大きく、画面に一度に収まらない場合、`list` コマンドは `less` コマンド（利用可能な場合）を使用して出力をページング（スクロール表示）します。この動作は `$NOTES_CLI_PAGER` でカスタマイズできます。

メモが大量にあると多くの行が出力されます。その場合、`less` のようなページャーツールが便利です。パイプを使って明示的に `less` に渡すことで、ページごとに出力を確認することもできます。グローバルオプションの `-A` は `--always-color` の略です。

```
$ notes -A ls --full | less -R

```

すべてのメモをすばやく確認したい場合は、`--full` よりも `--oneline`（または `-o`）の方が便利な場合があります。`notes ls --oneline` は、1つのメモの概要を1行で表示します。

出力例：

```
blog/how-to-handle-files.md golang,file How to handle files in Go

```

* 第1フィールドは、ホームディレクトリからのメモファイルの相対パスを異なる色で示します。パスの最初の部分（カテゴリ）は緑色、2番目の部分（ファイル名）は黄色で表示されます。
* 第2フィールドは、メモのタグをカンマ区切りで示します。メモにタグがない場合は空欄になります。
* 第3フィールドは、メモのタイトルです。

これは、多くのメモを一目で確認するのに便利です。出力が大きい場合は、利用可能であれば `less` がページングに使用されます。

詳細については、`notes list --help` を参照してください。

### メモのテンプレート

各カテゴリのディレクトリ、またはルートディレクトリにメモのテンプレートを作成できます。カテゴリディレクトリまたはホームに `.template.md` ファイルが配置されている場合、`notes new` を実行したときにその内容が自動的に挿入されます。

たとえば、以下の内容で `HOME/minutes/.template.md` を作成したとします。

```markdown
---

- アジェンダ:
- 出席者:



```

この状態で `notes new minutes weekly-meeting-2018-11-07` を実行すると、以下のようにテンプレートが挿入された新しいメモが作成されます。

```markdown
weekly-meeting-2018-11-07
=========================
- Category: minutes
- Tags:
- Created: 2018-11-07T14:19:27+09:00

---

- アジェンダ:
- 出席者:

```

カテゴリディレクトリにあるテンプレートファイルが優先されます。たとえば、以下のような配置になっている状況で `notes new minutes weekly-meeting-2018-11-07` を実行した場合、

```
HOME
├── .template.md
└── minutes
    └── .template.md

```

`HOME/.template.md` ではなく、`HOME/minutes/.template.md` が使用されます。

### Gitリポジトリにメモを保存する

最後に、メモをGitリポジトリのリビジョンとして保存することができます。

```
$ notes save

```

これにより、`notes-cli` ディレクトリ以下のすべてのメモがGitリポジトリとして保存されます。メモのすべての変更がステージング（add）され、自動的にコミットが作成されます。

デフォルトでは、リポジトリへの追加とコミットのみを行います。ただし、リポジトリにリモート（`origin`）を設定している場合は、自動的にリモートへメモがプッシュ（push）されます。

詳細については、`notes save --help` を参照してください。

### 環境変数で動作を設定する

前述の通り、いくつかの動作は環境変数で設定可能です。以下は `notes` の動作に影響を与えるすべての環境変数の一覧です。`$NOTES_` で始まる変数は `notes` コマンド専用のものです。その他は `notes` の動作に影響を与える一般的な環境変数です。
Git、エディタ、またはページャーとの連携を無効にしたい場合は、`export NOTES_CLI_PAGER=` のように、対応する環境変数に空文字列を設定してください。

| 変数名 | デフォルト値 | 説明 |
| --- | --- | --- |
| `$NOTES_CLI_HOME` | [XDG data dir](https://wiki.archlinux.org/index.php/XDG_Base_Directory) 配下の `notes-cli` | `notes` のホームディレクトリ。すべてのメモはこのサブディレクトリ内に保存されます。 |
| `$NOTES_CLI_EDITOR` | なし | お気に入りのエディタコマンド。`"vim -g"` のようなオプションを含めることができます。 |
| `$NOTES_CLI_GIT` | `"git"` | Gitコマンドのパス。メモをGitリポジトリとして保存するために使用されます。 |
| `$NOTES_CLI_PAGER` | `"less -R -F -X"` | `notes list` からの長い出力をページングするためのページャーコマンド。 |
| `$XDG_DATA_HOME` | なし | `$NOTES_CLI_HOME` が設定されていない場合、ホームとして使用されます。 |
| `$APPLOCALDATA` | なし | Windows環境において `$XDG_DATA_HOME` が設定されていない場合でも、ホームとして使用されます。 |
| `$EDITOR` | なし | `$NOTES_CLI_EDITOR` が設定されていない場合、エディタコマンドを選択するために参照されます。 |
| `$PAGER` | なし | `$NOTES_CLI_PAGER` が設定されていない場合、ページャーコマンドを選択するために参照されます。 |

設定内容は `notes config` コマンドで確認できます。

### 新しいサブコマンドを追加して `notes` コマンドを拡張する

拡張可能です。[Git](https://git-scm.com/) と同様に、`notes` コマンドはユーザーが未定義のサブコマンドを指定した際、外部のサブコマンドを実行しようと試みます。たとえば、`notes foo` と入力すると、`notes` コマンドはそれが組み込みのサブコマンドではないことを認識します。そして、同じ引数を使って `notes-foo` を実行しようとします。

実行される外部サブコマンドには、以下の引数が渡されます。

```
{notesへのフルパス} {グローバルオプション...} {サブコマンド} {ローカルオプション...}

```

たとえば、あなたの `$PATH` が通っている場所に、`notes-hello` という名前で以下のスクリプトが置かれているとします。

```sh
#!/bin/sh
echo "Hello! $*"

```

ここで `notes hello` を実行すると、指定した引数 `hello` が、`notes` のフルパスとともに実行された外部サブコマンドへそのまま渡されるため、`Hello! /path/to/bin/notes hello` と出力されます。
したがって、`notes --no-color hello --foo` を実行した場合は、`Hello! /path/to/bin/notes --no-color hello --foo` と出力されます。
すべての引数が転送されるため、サブコマンド側で、サブコマンドの前に指定されたグローバルオプションを参照することができます。

この外部サブコマンドのサポートは、自分の用途に合わせて `notes` の機能を拡張したい場合に便利です。たとえば以下のような使い方ができます。

* ブログのメモをブログサービスにアップロードするための独自コマンドを作成する。
* `ls -o -s modified` を `lsmod` のように呼び出す独自のエイリアスコマンドを作成する。

### シェルでの補完

* **zsh の場合：**

補完スクリプト `_notes` を、お使いの補完ディレクトリに配置してください。

```
$ git clone https://github.com/rhysd/notes-cli.git
$ cp nodes-cli/completions/zsh/_notes /path/to/completion/dir/

```

配置する補完ディレクトリは `$fpath` に登録されている必要があります。

```
fpath=(/path/to/completion/dir $fpath)

```

* **bash の場合：**

お使いの `.bashrc` に以下の行を追加してください。

```
$ eval "$(notes --completion-script-bash)"

```

* **fish の場合：**

`completions/fish/` 配下にある補完スクリプトを、お使いの補完ディレクトリにコピーしてください。

```
$ git clone https://github.com/rhysd/notes-cli.git
$ cp nodes-cli/completions/fish/notes.fish ~/.config/fish/completions/

```

### man マニュアルの設定

`notes` コマンドは `man` マニュアルファイルを生成することができます。

```
$ notes --help-man > /usr/local/share/man/man1/notes.1

```



### Goプログラムから使用する

このコマンドは、Goプログラムからライブラリとして使用することができます。インターフェースの詳細については、[APIドキュメント](http://godoc.org/github.com/rhysd/notes-cli) をお読みください。

## FAQ

### `/path/to/dir` のような任意のパスをホームに指定できますか？

環境変数にそのパスを設定してください。

```sh
export NOTES_CLI_HOME=/path/to/dir

```

### メモを grep するにはどうすればよいですか？

コマンドライン上で、`notes list` と grep ツールを組み合わせてください。たとえば、以下のようになります。

```sh
$ grep -E some word $(notes list)
$ ag some word $(notes list)

```

カテゴリやタグでフィルタリングしたい場合は、`list` コマンドの `-c` や `-t` オプションを使用してください。

### 対話的にメモをフィルタリングしてエディタで開くにはどうすればよいですか？

`notes list` からのパスのリストをパイプで渡してください。以下は `peco` と Vim を使用した例です。

```sh
$ notes list | peco | xargs -o vim --not-a-term

```

### リストから選択せずに、最新のメモを開くことはできますか？

`notes list` の出力は、デフォルトで作成日時の順にソートされています。`head` コマンドを使用することで、リスト内の最新のメモを選択できます。

```sh
$ vim "$(notes list | head -1)"

```

最後に変更されたメモにアクセスしたい場合は、`modified`（更新日時）でソートし、`head` で最初の項目を取得すれば機能します。

```sh
$ vim "$(notes list --sort modified | head -1)"

```

`notes list` に `--sort`（または `-s`）オプションを渡すことで、ソート方法を変更できます。詳細については `notes list --help` を参照してください。

### メモを削除するにはどうすればよいですか？

`rm` と `notes list` を使用してください。以下は、特定のカテゴリ `foo` のすべてのメモを削除する例です。

```sh
$ rm $(notes list -c foo)

```

Gitリポジトリ機能のおかげで、次に `notes save` を実行するまでは、メモが完全に削除されることはありません。

### メモ内にメタデータを表示したくありません。隠すことはできますか？

以下のように、メタデータをコメントアウトすることができます。

```markdown
some title
==========
本文

```

閉じコメント `-->` はメモの本文には含まれません。コメントアウトされたメタデータは（Markdownとしては）レンダリングされず、`notes` コマンドからのみ読み取られます。

### デフォルトでメタデータを隠すことはできますか？

可能です。`.template.md` が `-->`（閉じコメント）で始まっている場合、`notes` はユーザーがメタデータを隠すことを望んでいると自動的に判断し、適切な位置に ````

`notes new` を実行すると、以下のように新しいメモが作成されます。

```markdown
some-title
==========

```

### 画像などのリソースはどのように管理されますか？

ホームディレクトリの下にリソース用のディレクトリを作成することをお勧めします。

Markdown 以外のすべてのリソースは `notes` コマンドによって無視されます。そのため、メモの Markdown ファイルと同じディレクトリに `.png` や `.jpg` ファイルを自由に配置できます。

あるいは、`HOME/images/` や `HOME/category1/images` のように、画像専用の独立したディレクトリを使用することもできます。`grep` を使用する際、同じディレクトリ内に大量の画像とメモファイルが混在しているよりも、この方法の方が適している場合があります。

画像ディレクトリを他のカテゴリディレクトリと区別したい場合は、カテゴリディレクトリの名前に `.` プレフィックスを付けることができない特性を利用して、`HOME/.images` のように `.` プレフィックスを付けてください。

### デフォルトで `--color-always` を使用することは可能ですか？

以下のように、シェルのエイリアス機能を使用してください。

```sh
alias notes='notes --color-always'

```

### [memolist.vim](https://github.com/glidenote/memolist.vim) から移行するにはどうすればよいですか？

[移行スクリプト](./scripts/migrate-from-memolist.rb) を試してみてください。

```
$ git clone [https://github.com/rhysd/notes-cli.git](https://github.com/rhysd/notes-cli.git)
$ cd ./notes-cli
$ ruby ./scripts/migrate-from-memolist.rb /path/to/memolist/dir /path/to/note-cli/home

```

### Vim と統合（連携）するにはどうすればよいですか？

[notes-cli 用の Vim プラグイン](https://github.com/rhysd/vim-notes-cli) を試すことができます。

プラグインを入れるのが大げさだと感じる場合は、以下の設定を試すこともできます。お使いの `.vimrc` に以下のコードを記述してください。

```vim
function! s:notes_grep(args) abort
    let idx = match(a:args, '\s\+\ze/[^/]\+/')
    if idx <= 0
        " :NotesGrep /pat/ の場合
        let paths = join(split(system('notes list'), '\n'), ' ')
        execute 'vimgrep' a:args paths
        return
    endif

    " :NotesGrep {args} /pat/ の場合
    let paths = join(split(system('notes list ' . a:args[:idx]), '\n'), ' ')
    if paths ==# ''
        echohl ErrorMsg | echo 'No file found' | echohl None
        return
    endif
    let pat = a:args[idx:]
    execute 'vimgrep' pat paths
endfunction
command! -nargs=+ NotesGrep call <SID>notes_grep(<q-args>)

function! s:notes_new(...) abort
    if has_key(a:, 1)
        let cat = a:1
    else
        let cat = input('category?: ')
    endif
    if has_key(a:, 2)
        let name = a:2
    else
        let name = input('filename?: ')
    endif
    let tags = get(a:, 3, '')
    let cmd = printf('notes new --no-inline-input %s %s %s', cat, name, tags)
    let out = system(cmd)
    if v:shell_error
        echohl ErrorMsg | echomsg string(cmd) . ' failed: ' . out | echohl None
        return
    endif
    let path = split(out)[-1]
    execute 'edit!' path
    normal! Go
endfunction
command! -nargs=* NotesNew call <SID>notes_new(<f-args>)

function s:notes_last_mod(args) abort
    let out = system('notes list --sort modified ' . a:args)
    if v:shell_error
        echohl ErrorMsg | echomsg string(cmd) . ' failed: ' . out | echohl None
        return
    endif
    let last = split(out)[0]
    execute 'edit!' last
endfunction
command! -nargs=* NotesLastMod call <SID>notes_last_mod(<q-args>)

```

* `:NotesGrep [args] /pat/`: 指定された `/pat/` を使って `:vimgrep` でメモを検索します。`:vimgrep` のおかげで、検索結果はクイックフィックスリストに保存されます。`:copen` でクイックフィックスウィンドウを開くことで、一致した箇所を簡単に確認してリストからファイルを開くことができます。
* `:NotesNew [args]`: 新しいメモを作成し、新しいバッファで開きます。`args` は `notes new` と同じですが、カテゴリとファイル名は空にすることができます。その場合、コマンドの開始後に Vim から入力を求められます。
* `:NotesLastMod [args]`: 最後に変更されたメモを新しいバッファで開きます。`args` が指定された場合は、内部の `notes list` コマンドの実行に渡されるため、`-c` や `-t` を使ってカテゴリやタグで結果をフィルタリングできます。


## update
```sh
go get -u ./...
go mod tidy

# flakeのバージョンを更新
# go mod vendor
git add go.mod go.sum
nix build .

# flake.nixのvendor hashの更新
```

## ライセンス

[MIT License](LICENSE.txt)
