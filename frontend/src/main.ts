import { mount } from 'svelte'
import './style.css'
import App from './App.svelte'
import { suppressNativeMenu } from './lib/nativeMenu'

suppressNativeMenu(window)

const app = mount(App, { target: document.getElementById('app')! })

export default app
