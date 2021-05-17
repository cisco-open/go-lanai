const {merge} = require('webpack-merge');
const base = require('./webpack.base.js');
const path = require('path');
const CompressionPlugin = require("compression-webpack-plugin");

module.exports = env => merge(base(env), {
    mode: 'production',
    output: {
        path: path.resolve(env.OUTPUT_DIR),
    },
    plugins: [
        new CompressionPlugin({
            include: /.+\.js/,
            deleteOriginalAssets: true,
        }),
    ],
    optimization: {
        minimize: true,
    }
});