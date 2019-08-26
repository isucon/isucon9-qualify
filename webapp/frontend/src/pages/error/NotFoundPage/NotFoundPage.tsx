import React from 'react';
import makeStyles from '@material-ui/core/styles/makeStyles';
import { MuiThemeProvider, Theme } from '@material-ui/core';
import { themeInstance } from '../../../theme';
import Container from '@material-ui/core/Container';
import Typography from '@material-ui/core/Typography';
import { Link } from 'react-router-dom';
import { routes } from '../../../routes/Route';

const useStyles = makeStyles((theme: Theme) => ({
  container: {
    paddingTop: theme.spacing(2),
    display: 'flex',
    flexDirection: 'column',
    alignItems: 'center',
  },
  img: {
    width: '70%',
  },
  message: {
    paddingTop: theme.spacing(1),
  },
  link: {
    paddingTop: theme.spacing(2),
  },
}));

export type Props = {
  message?: string;
};

const NotFoundPage: React.FC<Props> = ({ message }) => {
  const classes = useStyles();

  return (
    <MuiThemeProvider theme={themeInstance}>
      <Container maxWidth="md" className={classes.container}>
        <img className={classes.img} src={'/not_found.png'} alt={'not found'} />
        <Typography variant="h3">404 Not Found</Typography>
        {message && (
          <Typography variant="h4" className={classes.message}>
            {message}
          </Typography>
        )}
        <Link to={routes.top.path}>
          <Typography variant="h6" className={classes.link}>
            トップページへ
          </Typography>
        </Link>
      </Container>
    </MuiThemeProvider>
  );
};

export { NotFoundPage };
