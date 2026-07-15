import { viteLogo } from './src/core/config'
import Banner from 'vite-plugin-banner'
import * as path from 'path'
import { loadEnv } from 'vite'
import vuePlugin from '@vitejs/plugin-vue'
import vueDevTools from 'vite-plugin-vue-devtools'
import VueFilePathPlugin from './vitePlugin/componentName/index.js'
import { svgBuilder } from './vitePlugin/svgBuilder/index.js'
import UnoCSS from '@unocss/vite'

// @see https://cn.vitejs.dev/config/
export default ({ mode }) => {
  const env = loadEnv(mode, process.cwd())
  viteLogo(env)

  const timestamp = Date.parse(new Date())

  const alias = {
    '@': path.resolve(import.meta.dirname, './src'),
    vue$: 'vue/dist/vue.runtime.esm-bundler.js'
  }

  const base = '/'
  const root = './'
  const outDir = 'dist'

  const config = {
    base: base, // 编译后js导入的资源路径
    root: root, // index.html文件所在位置
    publicDir: 'public', // 静态资源文件夹
    resolve: {
      alias
    },
    css: {
      preprocessorOptions: {
        scss: {
          api: 'modern-compiler' // or "modern"
        }
      }
    },
    server: {
      // 如果使用docker-compose开发模式，设置为false
      open: true,
      port: Number(env.VITE_CLI_PORT),
      proxy: {
        // 把key的路径代理到target位置
        // detail: https://cli.vuejs.org/config/#devserver-proxy
        [env.VITE_BASE_API]: {
          // 需要代理的路径   例如 '/api'
          target: `${env.VITE_BASE_PATH}:${env.VITE_SERVER_PORT}/`, // 代理到 目标路径
          changeOrigin: true,
          rewrite: (path) =>
            path.replace(new RegExp('^' + env.VITE_BASE_API), '')
        }
      }
    },
    build: {
      manifest: false, // 是否产出manifest.json
      sourcemap: false, // 是否产出sourcemap.json
      outDir: outDir, // 产出目录
      target: 'es2015',
      rolldownOptions: {
        output: {
          entryFileNames: 'assets/087AC4D233B64EB0[name].[hash].js',
          chunkFileNames: 'assets/087AC4D233B64EB0[name].[hash].js',
          assetFileNames: 'assets/087AC4D233B64EB0[name].[hash].[ext]'
        }
      }
    },
    plugins: [
      env.VITE_POSITION === 'open' &&
      vueDevTools({ launchEditor: env.VITE_EDITOR }),
      vuePlugin(),
      svgBuilder(['./src/plugin/', './src/assets/icons/'], base, outDir, 'assets', mode),
      [Banner(`\n Build based on asset-center \n Time : ${timestamp}`)],
      VueFilePathPlugin('./src/pathInfo.json'),
      UnoCSS()
    ]
  }
  return config
}
