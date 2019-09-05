import * as React from 'react';
import makeStyles from '@material-ui/core/styles/makeStyles';
import { AppBar, MuiThemeProvider, Theme } from '@material-ui/core';
import Drawer from '@material-ui/core/Drawer';
import List from '@material-ui/core/List';
import ListItem from '@material-ui/core/ListItem';
import ListItemText from '@material-ui/core/ListItemText';
import Toolbar from '@material-ui/core/Toolbar';
import IconButton from '@material-ui/core/IconButton';
import MenuIcon from '@material-ui/icons/Menu';
import Typography from '@material-ui/core/Typography';
import { ExpandLess, ExpandMore } from '@material-ui/icons';
import Collapse from '@material-ui/core/Collapse';
import { CategorySimple } from '../../dataObjects/category';
import NewReleasesIcon from '@material-ui/icons/NewReleases';
import ListItemIcon from '@material-ui/core/ListItemIcon';
import EventSeatIcon from '@material-ui/icons/EventSeat';
import SettingsIcon from '@material-ui/icons/Settings';
import PersonIcon from '@material-ui/icons/Person';
import WeekendIcon from '@material-ui/icons/Weekend';
import { themeInstance } from '../../theme';

const useStyles = makeStyles((theme: Theme) => ({
  appBar: {
    color: theme.palette.primary.main,
    backgroundColor: theme.palette.primary.contrastText,
  },
  text: {
    fontWeight: theme.typography.fontWeightBold,
    textAlign: 'center',
    width: '100%', // センタリング
    cursor: 'pointer',
  },
  list: {
    width: '300px',
  },
  nested: {
    paddingLeft: theme.spacing(4),
  },
}));

interface Props {
  isLoggedIn: boolean;
  ownUserId: number;
  categories: CategorySimple[];
  goToTopPage: () => void;
  goToUserPage: (userId: number) => void;
  goToSettingPage: () => void;
  goToCategoryItemList: (categoryId: number) => void;
  onClickTitle: (isLoggedIn: boolean) => void;
}

const Header: React.FC<Props> = ({
  isLoggedIn,
  ownUserId,
  categories,
  goToTopPage,
  goToUserPage,
  goToSettingPage,
  goToCategoryItemList,
  onClickTitle,
}) => {
  const classes = useStyles();
  const [state, setState] = React.useState({
    open: false,
    categoryExpanded: false,
  });

  const { open, categoryExpanded } = state;

  const onClickNewItems = (e: React.MouseEvent) => {
    e.preventDefault();
    goToTopPage();
  };

  const onExpandCategory = (e: React.MouseEvent) => {
    e.preventDefault();
    setState({ ...state, categoryExpanded: !state.categoryExpanded });
  };

  const onClickCategory = (categoryId: number) => (e: React.MouseEvent) => {
    e.preventDefault();
    goToCategoryItemList(categoryId);
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

  const onClickTitleText = (e: React.MouseEvent) => {
    e.preventDefault();
    onClickTitle(isLoggedIn);
  };

  // MEMO: Wrap component by MuiThemeProvider again to ignore this bug. https://github.com/mui-org/material-ui/issues/14044
  return (
    <MuiThemeProvider theme={themeInstance}>
      {isLoggedIn && (
        <Drawer open={open} onClose={toggleDrawer(false)}>
          <List className={classes.list}>
            <ListItem button onClick={onClickNewItems}>
              <ListItemIcon>
                <NewReleasesIcon color="primary" />
              </ListItemIcon>
              <ListItemText primary="新着商品" />
            </ListItem>
            <ListItem button onClick={onExpandCategory}>
              <ListItemIcon>
                <WeekendIcon color="primary" />
              </ListItemIcon>
              <ListItemText primary="カテゴリタイムライン" />
              {categoryExpanded ? <ExpandLess /> : <ExpandMore />}
            </ListItem>
            <Collapse in={categoryExpanded} timeout="auto" unmountOnExit>
              <List disablePadding>
                {categories.map((category: CategorySimple) => (
                  <ListItem
                    button
                    onClick={onClickCategory(category.id)}
                    className={classes.nested}
                  >
                    <ListItemIcon>
                      <EventSeatIcon color="primary" />
                    </ListItemIcon>
                    <ListItemText primary={category.categoryName} />
                  </ListItem>
                ))}
              </List>
            </Collapse>
            <ListItem button onClick={onClickMyPage}>
              <ListItemIcon>
                <PersonIcon color="primary" />
              </ListItemIcon>
              <ListItemText primary="マイページ" />
            </ListItem>
            <ListItem button onClick={onClickMySettingPage}>
              <ListItemIcon>
                <SettingsIcon color="primary" />
              </ListItemIcon>
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
          <Typography
            className={classes.text}
            variant="h5"
            onClick={onClickTitleText}
          >
            ISUCARI
          </Typography>
        </Toolbar>
      </AppBar>
    </MuiThemeProvider>
  );
};

export { Header };
