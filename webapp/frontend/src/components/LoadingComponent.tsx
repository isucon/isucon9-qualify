import React from 'react';
import { makeStyles } from '@material-ui/core/styles';
import CircularProgress from '@material-ui/core/CircularProgress';

const useStyles = makeStyles(theme => ({
    progress: {
        top: 0,
        bottom: 0,
        right: 0,
        left: 0,
        margin: 'auto',
        position: 'absolute',
    },
}));

const LoadingComponent: React.FC = () => {
    const classes = useStyles();

    return (
        <React.Fragment>
            <CircularProgress
                size={80}
                className={classes.progress}
            />
        </React.Fragment>
    );
};

export default LoadingComponent;
