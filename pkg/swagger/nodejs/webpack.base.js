const path = require('path');
const { ProvidePlugin } = require('webpack');
const CopyPlugin = require("copy-webpack-plugin");
const TerserPlugin = require("terser-webpack-plugin");

module.exports = env => {
  return {
    entry: {
      "swagger-ui-bundle": {
        import: ['swagger-ui', 'swagger-ui/dist/swagger-ui-standalone-preset']
      },
      "nfv-swagger-ui": {
        import: './js/nfv-swagger-ui.js',
        dependOn: ['swagger-ui-bundle'],
      },
    },
    output: {
      filename: 'js/[name].js',
    },
    performance: {
      maxAssetSize: 512000,
      maxEntrypointSize: 512000,
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
      new ProvidePlugin({
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
    ],
    optimization: {
      minimize: true,
      minimizer: [
        new TerserPlugin({
          include: [/nfv-swagger-ui\.js/],
          extractComments: false,
        }),
      ],
    },
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