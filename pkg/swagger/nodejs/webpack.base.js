/**
 * Copyright 2023 Cisco Systems, Inc. and its affiliates
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

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
            loader: '@svgr/webpack'
          }
        }
      ]
    }
  }
};