import React from 'react'
import { HeaderProps } from './Header.types'
import styles from './Header.module.css'

const Header = ({title}: HeaderProps) => {
  return (
    <header className={styles.header}>
        <div>logo</div>
        <h1>{title}</h1>
        <button>бургер</button>
    </header>
  )
}

export default Header