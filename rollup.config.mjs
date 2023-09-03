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
      sourcemapIgnoreList: (relativeSourcePath, sourcemapPath) => {
        // will ignore-list all files with node_modules in their paths
        return relativeSourcePath.includes('node_modules');
      },
      plugins: [terser()]
    }
  ],
  plugins: [nodeResolve()]
};
