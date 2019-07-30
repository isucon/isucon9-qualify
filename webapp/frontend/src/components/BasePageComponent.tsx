import React from 'react';

import { Container, MuiThemeProvider, createMuiTheme } from '@material-ui/core';

const themeInstance = createMuiTheme({
    palette: {
        background: {
            default: 'white'
        },
    },
});

const BasePageComponent: React.FC = ({children}) => (
    <MuiThemeProvider theme={themeInstance}>
        <Container maxWidth="lg" children={children} />
    </MuiThemeProvider>
);

export { BasePageComponent }