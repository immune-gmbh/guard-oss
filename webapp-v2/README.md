# Basic Readme For Now:

- Node v16.3
- npm install
- npm run dev

## VS Code:

- https://marketplace.visualstudio.com/items?itemName=bradlc.vscode-tailwindcss

## Generate AuthSrv Typescript Schema
```
cd authsrv2
bundle exec schema2type -o ../webapp-v2/generated/authsrvSchema.d.ts 
```
After that, `export` the schema namespace.
## Generate AuthSrv Routes 
```
cd authsrv2
bundle exec rake js_routes_rails:export
```
