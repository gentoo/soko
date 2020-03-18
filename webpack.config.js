const webpack = require('webpack');
const path = require('path');

module.exports = {
    entry: {
        stylesheets: './web/packs/stylesheets.js',
        application: './web/packs/application.js',
        index: './web/packs/index.js',
        packages: './web/packs/packages.js',
        useflags: './web/packs/useflags.js',
    },
    output: {
        path: path.resolve(__dirname, 'assets'),
        filename: '[name].js',
    },
    plugins: [
        require('postcss-import')
    ],
    module: {
        rules: [
            {
                test: /\.s[ac]ss$/i,
                use: [
                    // Creates `style` nodes from JS strings
                    'style-loader',
                    // Translates CSS into CommonJS
                    {
                        loader: 'css-loader',
                    },{
                        loader: 'resolve-url-loader',
                    },
                    // Compiles Sass to CSS
                    {
                        loader: 'sass-loader',
                        options: {
                            sourceMap: true,
                        }
                    },
                ],
            },
            {
                test: /\.(woff(2)?|ttf|eot|svg)(\?v=\d+\.\d+\.\d+)?$/,
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
    ],
};
