import React from 'react';
import makeStyles from '@material-ui/core/styles/makeStyles';
import SignUpFormContainer from '../containers/SignUpFormContainer';
import BasePageContainer from '../containers/BasePageContainer';
import { Theme } from '@material-ui/core';

const useStyles = makeStyles((theme: Theme) => ({
  paper: {
    marginTop: theme.spacing(1),
    display: 'flex',
    flexDirection: 'column',
    alignItems: 'center',
  },
}));

const SignUpPage: React.FC = () => {
  const classes = useStyles();

  return (
    <BasePageContainer>
      <div className={classes.paper}>
        <SignUpFormContainer />
      </div>
    </BasePageContainer>
  );
};

export default SignUpPage;
