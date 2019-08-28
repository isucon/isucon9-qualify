import * as React from 'react';
import makeStyles from '@material-ui/core/styles/makeStyles';
import { AppBar, Theme } from '@material-ui/core';
import Drawer from '@material-ui/core/Drawer';
import List from '@material-ui/core/List';
import ListItem from '@material-ui/core/ListItem';
import ListItemText from '@material-ui/core/ListItemText';
import Toolbar from '@material-ui/core/Toolbar';
import IconButton from '@material-ui/core/IconButton';
import MenuIcon from '@material-ui/icons/Menu';
import Typography from '@material-ui/core/Typography';

const useStyles = makeStyles((theme: Theme) => ({
  appBar: {
    color: theme.palette.primary.main,
    backgroundColor: theme.palette.primary.contrastText,
  },
  text: {
    fontWeight: theme.typography.fontWeightBold,
    textAlign: 'center',
    width: '100%', // センタリング
  },
  list: {
    width: '200px',
  },
}));

interface Props {
  isLoggedIn: boolean;
  ownUserId: number;
  goToTopPage: () => void;
  goToUserPage: (userId: number) => void;
  goToSettingPage: () => void;
}

const Header: React.FC<Props> = ({
  isLoggedIn,
  ownUserId,
  goToTopPage,
  goToUserPage,
  goToSettingPage,
}) => {
  const classes = useStyles();
  const [state, setState] = React.useState({
    open: false,
  });

  const onClickTop = (e: React.MouseEvent) => {
    e.preventDefault();
    goToTopPage();
  };

  const onClickMyPage = (e: React.MouseEvent) => {
    e.preventDefault();
    goToUserPage(ownUserId);
  };

  const onClickMySettingPage = (e: React.MouseEvent) => {
    e.preventDefault();
    goToSettingPage();
  };

  const toggleDrawer = (open: boolean) => (event: React.MouseEvent) => {
    event.preventDefault();
    setState({ ...state, open });
  };

  return (
    <React.Fragment>
      {isLoggedIn && (
        <Drawer open={state.open} onClose={toggleDrawer(false)}>
          <List className={classes.list}>
            <ListItem button onClick={onClickTop}>
              <ListItemText primary="新着商品" />
            </ListItem>
            <ListItem button onClick={onClickMyPage}>
              <ListItemText primary="マイページ" />
            </ListItem>
            <ListItem button onClick={onClickMySettingPage}>
              <ListItemText primary="設定" />
            </ListItem>
          </List>
        </Drawer>
      )}
      <AppBar className={classes.appBar} position="fixed">
        <Toolbar>
          {isLoggedIn && (
            <IconButton
              color="inherit"
              onClick={toggleDrawer(true)}
              edge="start"
            >
              <MenuIcon />
            </IconButton>
          )}
          <Typography className={classes.text} variant="h5" noWrap>
            ISUCARI
          </Typography>
        </Toolbar>
      </AppBar>
    </React.Fragment>
  );
};

export { Header };
