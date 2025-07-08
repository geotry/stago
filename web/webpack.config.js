const path = require("path");
const HtmlWebpackPlugin = require('html-webpack-plugin');

module.exports = {
  entry: {
    main: './src/main.js',
    worker: './src/worker.js',
  },
  output: {
    filename: '[name].bundle.js',
    path: path.resolve(__dirname, 'dist'),
    clean: true,
  },
  plugins: [
    new HtmlWebpackPlugin({
      template: path.resolve(__dirname, 'index.html'),
    })
  ],
  optimization: {
    runtimeChunk: 'single',
  },
  devServer: {
    static: [
      {
        directory: path.join(__dirname, "shaders"),
        publicPath: "/shaders",
      },
    ],
    client: {
      overlay: true,
    },
    compress: true,
    hot: true,
    port: 8080,
  }
};