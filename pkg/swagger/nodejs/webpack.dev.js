const {merge} = require('webpack-merge');
const base = require('./webpack.base.js');
const path = require('path');
const BundleAnalyzerPlugin = require('webpack-bundle-analyzer').BundleAnalyzerPlugin;

module.exports = env => merge(base(env), {
    mode: 'development',
    output: {
        path: path.resolve(__dirname, 'dist'),
    },
    plugins: [
        new BundleAnalyzerPlugin({
            analyzerMode: "static",
            openAnalyzer: false,
            excludeAssets: [/swagger-ui-bundle\.js/],
            reportFilename: "./webpack.bundle.report.html"
        }),
    ],
    optimization: {
        minimize: false,
    }
});