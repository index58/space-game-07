declare module "node:fs" {
  export function existsSync(path: string | URL): boolean;
  export function readdirSync(path: string | URL): string[];
  export function readFileSync(path: string | URL, encoding: "utf8"): string;
  export function statSync(path: string | URL): {
    // Показывает, что путь указывает на директорию.
    isDirectory(): boolean;
  };
}
