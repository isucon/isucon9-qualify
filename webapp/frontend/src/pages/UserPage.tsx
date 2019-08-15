import React, { ReactElement } from 'react';
import { ItemListComponent } from '../components/ItemListComponent';
import { ItemData, TransactionItem } from '../dataObjects/item';
import { UserData } from '../dataObjects/user';
import Avatar from '@material-ui/core/Avatar';
import { createStyles, Grid, Theme, WithStyles } from '@material-ui/core';
import Typography from '@material-ui/core/Typography';
import Divider from '@material-ui/core/Divider';
import SellingButtonContainer from '../containers/SellingButtonContainer';
import BasePageContainer from '../containers/BasePageContainer';
import { ErrorProps, PageComponentWithError } from '../hoc/withBaseComponent';
import { RouteComponentProps } from 'react-router';
import LoadingComponent from '../components/LoadingComponent';
import { TransactionListComponent } from '../components/TransactionListComponent';
import AppBar from '@material-ui/core/AppBar/AppBar';
import Tabs from '@material-ui/core/Tabs/Tabs';
import Tab from '@material-ui/core/Tab/Tab';
import { StyleRules } from '@material-ui/core/styles';
import withStyles from '@material-ui/core/styles/withStyles';

const styles = (theme: Theme): StyleRules =>
  createStyles({
    avatar: {
      margin: theme.spacing(3),
      width: '100px',
      height: '100px',
    },
    tab: {
      color: theme.palette.text.primary,
      backgroundColor: theme.palette.primary.light,
    },
    list: {
      marginTop: theme.spacing(4),
    },
  });

interface UserPageProps extends WithStyles<typeof styles> {
  loading: boolean;
  load: (userId: number) => void;
  loggedInUserId: number;
  items: ItemData[];
  itemsHasNext: boolean;
  itemsLoadMore: (
    userId: number,
    itemId: number,
    createdAt: number,
    page: number,
  ) => void;
  transactions: TransactionItem[];
  transactionsHasNext: boolean;
  transactionsLoadMore: (
    itemId: number,
    createdAt: number,
    page: number,
  ) => void;
  user: UserData;
}

type Props = UserPageProps &
  ErrorProps &
  RouteComponentProps<{ user_id: string }>;

type State = {
  tabValue: any;
};

class UserPage extends React.Component<Props, State> {
  private ITEM_LIST_TAB = 0;
  private TRANSACTION_LIST_TAB = 1;

  constructor(props: Props) {
    super(props);

    this.props.load(Number(this.props.match.params.user_id));
    this.state = {
      tabValue: this.ITEM_LIST_TAB,
    };

    this._handleChange = this._handleChange.bind(this);
  }

  _handleChange(e: React.ChangeEvent<{}>, newValue: any) {
    this.setState({
      tabValue: newValue,
    });
  }

  _getItemList(): ReactElement {
    const { items } = this.props;

    if (items.length === 0) {
      return <Typography>現在出品されている商品はありません</Typography>;
    }

    const { itemsHasNext, itemsLoadMore, user } = this.props;
    const lastItem = items[items.length - 1];

    return (
      <ItemListComponent
        items={items}
        hasNext={itemsHasNext}
        loadMore={itemsLoadMore.bind(
          null,
          user.id,
          lastItem.id,
          lastItem.createdAt,
        )}
      />
    );
  }

  _getTransactionsList(): ReactElement {
    const { transactions } = this.props;

    if (transactions.length === 0) {
      return <Typography>取引はありません</Typography>;
    }

    const { transactionsHasNext, transactionsLoadMore } = this.props;
    const lastTransaction = transactions[transactions.length - 1];

    return (
      <TransactionListComponent
        items={transactions}
        hasNext={transactionsHasNext}
        loadMore={transactionsLoadMore.bind(
          null,
          lastTransaction.id,
          lastTransaction.createdAt,
        )}
      />
    );
  }

  render() {
    const { loading, user, loggedInUserId, classes } = this.props;
    const { tabValue } = this.state;
    const isMyPage = loggedInUserId === user.id;

    if (loading) {
      return <LoadingComponent />;
    }

    return (
      <BasePageContainer>
        <p>User Page</p>
        <Grid
          container
          direction="row"
          justify="center"
          alignItems="center"
          wrap="nowrap"
          spacing={2}
        >
          <Grid item>
            <Avatar className={classes.avatar}>
              {user.accountName.charAt(0)}
            </Avatar>
          </Grid>
          <Grid item xs>
            <Typography variant="h3">{user.accountName}</Typography>
          </Grid>
        </Grid>
        <Divider variant="middle" />
        <AppBar className={classes.tab}>
          <Tabs value={tabValue} onChange={this._handleChange}>
            <Tab label="出品商品" id="tab--item-list" />
            {isMyPage && <Tab label="取引一覧" id="tab--item-list" />}
          </Tabs>
        </AppBar>
        <div
          className={classes.list}
          id="tab--item-list"
          hidden={tabValue !== this.ITEM_LIST_TAB}
        >
          {this._getItemList()}
        </div>
        {isMyPage && (
          <div
            className={classes.list}
            id="tab--transactions-list"
            hidden={tabValue !== this.TRANSACTION_LIST_TAB}
          >
            {this._getTransactionsList()}
          </div>
        )}
        <SellingButtonContainer />
      </BasePageContainer>
    );
  }
}

export default PageComponentWithError<any>()(withStyles(styles)(UserPage));
