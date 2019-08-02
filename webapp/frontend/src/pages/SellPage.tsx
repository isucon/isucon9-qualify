import React from 'react';

import { makeStyles } from '@material-ui/core';
import SellFormContainer from "../containers/SellFormContainer";

const useStyles = makeStyles(theme => ({
    paper: {
        marginTop: theme.spacing(1),
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
    },
}));

const SellPage: React.FC = () => {
    const classes = useStyles();

    return (
        <div className={classes.paper}>
            <SellFormContainer />
        </div>
    );
};

export { SellPage }
