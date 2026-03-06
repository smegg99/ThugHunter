import { rm, cp } from "node:fs/promises";
import path from "node:path";

const dist = path.resolve("dist");
const pub = path.resolve(".output/public");

await rm(dist, { recursive: true, force: true });
await cp(pub, dist, { recursive: true });
console.log("Copied .output/public -> dist");
