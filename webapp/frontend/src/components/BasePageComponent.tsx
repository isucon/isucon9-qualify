import React from 'react';

import { Container } from '@material-ui/core';

const BasePageComponent: React.FC = ({children}) => (
    <Container maxWidth="xs" children={children} />
);

export { BasePageComponent }