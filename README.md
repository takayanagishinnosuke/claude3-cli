# CLAUDE3-CLI

# Use the CLI to interact with the CLAUDE3 API.

0. 前提条件

   - AWS Bedrock で claude-3-sonnet がデプロイされていること
   - AWS のアクセスキーが取得済みで必要なアクセス権が付与されていること

1. install

   ```
   brew install takayanagishinnosuke/tap/claude3-cli
   ```

   ※Homebrew 以外でのインストールはリリースページから環境に合ったバイナリをダウンロードして解凍してください。

2. 環境変数を設定する
   ```
   export AWS_ACCESS_KEY_ID=<AWS ACCESS KEY ID>
   export AWS_SECRET_ACCESS_KEY=<AWS SECRET ACCESS KEY>
   export AWS_DEFAULT_REGION=<Bedrok Deploy Region>
   ```
3. コマンドを実行する

   ```
   claude3-cli --help
   ```

   チャット

   ```
   claude3-cli
   ```

   画像の解析

   ```
   claude3-cli -img <FilePath>
   ```

4. Uninstall
   ```
   brew uninstall claude3-cli
   ```
