import React from 'react';
import { makeStyles } from '@material-ui/core/styles';
import CircularProgress from '@material-ui/core/CircularProgress';

const useStyles = makeStyles(theme => ({
    progress: {
        margin: theme.spacing(2),
    },
}));

const LoadingComponent: React.FC = () => {
    const classes = useStyles();

    return (
        <React.Fragment>
            <CircularProgress className={classes.progress}/>
        </React.Fragment>
    );
};

export default LoadingComponent;
