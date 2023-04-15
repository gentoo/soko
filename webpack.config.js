const MiniCssExtractPlugin = require('mini-css-extract-plugin');
const webpack = require('webpack');
const path = require('path');

module.exports = {
    entry: {
        stylesheets: './web/packs/stylesheets.js',
        application: './web/packs/application.js',
        graphiql: './web/packs/graphiql.js',
        index: './web/packs/index.js',
        packages: './web/packs/packages.js',
        useflags: './web/packs/useflags.js',
        userpref: './web/packs/userpref.js',
    },
    mode: 'production',
    output: {
        path: path.resolve(__dirname, 'assets'),
        filename: '[name].js',
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
                use: [
                    {
                        loader: 'file-loader',
                        options: {
                            name: '[name].[ext]',
                            publicPath: '/assets'
                        }
                    }
                ]
            }
        ],
    },
    plugins: [
        new webpack.ProvidePlugin({
            $: 'jquery',
            jQuery: 'jquery',
            'window.jQuery': 'jquery',
            'windows.jQuery': 'jquery',
            tether: 'tether',
            Tether: 'tether',
            'window.Tether': 'tether',
            Popper: ['popper.js', 'default'],
            'window.Tether': 'tether',
            Modal: 'exports-loader?Modal!bootstrap/js/dist/modal',
        }),
        new MiniCssExtractPlugin({
            filename: '[name].css',
        }),
    ],
};
