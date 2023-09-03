import { nodeResolve } from '@rollup/plugin-node-resolve';
import terser from '@rollup/plugin-terser';

export default {
  input: 'pkg/web/static/js/index.js',
  watch: {
    include: 'pkg/web/static/js/**'
  },
  output: [
    {
      file: 'pkg/web/static/js/index.min.js',
      format: 'es',
      sourcemap: true,
      sourcemapFile: 'pkg/web/static/js/index.min.js.map',
      plugins: [terser()]
    }
  ],
  plugins: [nodeResolve()]
};
