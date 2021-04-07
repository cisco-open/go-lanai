const path = require('path');
const CopyPlugin = require("copy-webpack-plugin");
const BundleAnalyzerPlugin = require('webpack-bundle-analyzer').BundleAnalyzerPlugin;

module.exports = env => {
  return {
    entry: {
      "nfv-swagger-ui": ['./js/nfv-swagger-ui.js'],
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
      ]
    },
    plugins: [
      new CopyPlugin([
          { from: "node_modules/swagger-ui/dist/swagger-ui.css", to: "css/"},
          { from: "node_modules/swagger-ui/dist/oauth2-redirect.html", to: "./"},
          { from: "css", to: "css"},
          { from: "fonts", to: "fonts"},
          { from: "images", to: "images"},
          { from: "*.html"},
      ]),
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