const webpack = require('webpack');
const path = require('path');
const CopyPlugin = require("copy-webpack-plugin");
const CompressionPlugin = require("compression-webpack-plugin");
const BundleAnalyzerPlugin = require('webpack-bundle-analyzer').BundleAnalyzerPlugin;

module.exports = env => {
  return {
    entry: {
      // "nfv-swagger-ui": ['./js/nfv-swagger-ui.js'],

      "swagger-ui-bundle": {
        import: ['swagger-ui', 'swagger-ui/dist/swagger-ui-standalone-preset']
      },
      "nfv-swagger-ui": {
        import: './js/nfv-swagger-ui.js',
        dependOn: ['swagger-ui-bundle'],
      },
    },
    output: {
      path: path.resolve(env.OUTPUT_DIR),
      filename: 'js/[name].js',
    },
    resolveLoader: {
      modules: [path.join(__dirname, "node_modules")],
    },
    resolve: {
      extensions: ['.js', '.jsx'],
      modules: [
        path.join(__dirname, "./js"),
        "node_modules"
      ],
      fallback: {
        "buffer": require.resolve("buffer/")
      },
    },
    plugins: [
      new webpack.ProvidePlugin({
        process: 'process/browser.js',
        Buffer: ['buffer', 'Buffer'],
      }),
      new CopyPlugin({
        patterns: [
          { from: "node_modules/swagger-ui/dist/swagger-ui.css", to: "css/"},
          { from: "node_modules/swagger-ui/dist/oauth2-redirect.html"},
          { from: "css", to: "css"},
          { from: "fonts", to: "fonts"},
          { from: "images", to: "images"},
          { from: "*.html"},
        ],
      }),
      new CompressionPlugin({
        include: /.+\.js/,
        deleteOriginalAssets: true,
      }),
      new BundleAnalyzerPlugin({
        analyzerMode: "static",
        openAnalyzer: false
      })
    ],
    module: {
      rules: [
        {
          test: /\.(js(x)?)(\?.*)?$/,
          exclude: /node_modules/,
          use: {
            loader: 'babel-loader'
          }
        },
        {
          test: /\.svg$/,
          exclude: /node_modules/,
          use: {
            loader: 'react-svg-loader'
          }
        }
      ]
    }
  }
};