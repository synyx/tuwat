import { nodeResolve } from '@rollup/plugin-node-resolve';

export default {
  input: 'pkg/web/static/js/index.js',
  output: {
      file: 'pkg/web/static/js/index.min.js',
      format: 'es',
      sourcemap: true,
      plugins: []
  },
  plugins: [nodeResolve()]
};
