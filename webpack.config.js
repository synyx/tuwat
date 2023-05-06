const path = require('path');

module.exports = {
    entry: './pkg/web/static/js/index.js',
    mode: "production",
    devtool: "source-map",
    output: {
        filename: 'main.js',
        path: path.resolve(__dirname, 'pkg/web/static/js'),
    },
};
