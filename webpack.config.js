const CopyPlugin = require('copy-webpack-plugin');
const MiniCssExtractPlugin = require('mini-css-extract-plugin');
const webpack = require('webpack');
const path = require('path');

module.exports = {
    entry: {
        stylesheets: './web/packs/stylesheets.js',
        application: './web/packs/application.js',
        index:       './web/packs/index.js',
        useflags:    './web/packs/useflags.js',
    },
    mode: 'production',
    externals: {
        jquery: 'jQuery',
    },
    output: {
        path: path.resolve(__dirname, 'assets'),
        filename: '[name].js',
        assetModuleFilename: '[name].[ext]'
    },
    module: {
        rules: [
            {
                test: /\.s[ac]ss$/i,
                use: [
                    MiniCssExtractPlugin.loader,
                    'css-loader',
                    'resolve-url-loader',
                    {
                        loader: 'sass-loader',
                        options: {
                            sourceMap: true,
                        }
                    },
                ],
            },
            {
                test: /\.(woff(2)?|ttf|eot|svg|png)(\?v=\d+\.\d+\.\d+)?$/,
                type: 'asset/resource',
            }
        ],
    },
    plugins: [
        new CopyPlugin({
            patterns: [
                {
                    from: require.resolve('jquery/dist/jquery.min.js'),
                    to: 'jquery.min.js',
                    info: { minimized: true },
                },
            ],
        }),
        new webpack.ProvidePlugin({
            $: 'jquery',
            jQuery: 'jquery',
            'window.jQuery': 'jquery',
            'windows.jQuery': 'jquery',
        }),
        new MiniCssExtractPlugin({
            filename: '[name].css',
        }),
    ],
};
