import React, { ReactElement } from 'react';
import { ItemList } from '../components/ItemList';
import { ItemData, TransactionItem } from '../dataObjects/item';
import { UserData } from '../dataObjects/user';
import Avatar from '@material-ui/core/Avatar';
import { createStyles, Theme, WithStyles } from '@material-ui/core';
import Grid from '@material-ui/core/Grid';
import Typography from '@material-ui/core/Typography';
import Divider from '@material-ui/core/Divider';
import SellingButtonContainer from '../containers/SellingButtonContainer';
import BasePageContainer from '../containers/BasePageContainer';
import { ErrorProps, PageComponentWithError } from '../hoc/withBaseComponent';
import { RouteComponentProps } from 'react-router';
import LoadingComponent from '../components/LoadingComponent';
import { TransactionList } from '../components/TransactionList';
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
    numSellItems: {
      marginTop: theme.spacing(1),
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
  load: (userId: number, isMyPage: boolean) => void;
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
  loading: boolean;
  currentPageUserId: number;
  isMyPage: boolean;
};

class UserPage extends React.Component<Props, State> {
  private ITEM_LIST_TAB = 0;
  private TRANSACTION_LIST_TAB = 1;

  constructor(props: Props) {
    super(props);

    const pageUserId = Number(this.props.match.params.user_id);
    const isMyPage = this.props.loggedInUserId === pageUserId;
    this.props.load(pageUserId, isMyPage);
    this.state = {
      tabValue: this.ITEM_LIST_TAB,
      loading: this.props.loading,
      currentPageUserId: pageUserId,
      isMyPage,
    };

    this._handleChange = this._handleChange.bind(this);
  }

  static getDerivedStateFromProps(nextProps: Props, prevState: State) {
    const nextLoading = nextProps.loading;
    const nextPageUserId = Number(nextProps.match.params.user_id);
    const nextIsMyPage = nextProps.loggedInUserId === nextPageUserId;

    // ページ遷移を確認した場合はデータ取得を行う
    if (nextPageUserId !== prevState.currentPageUserId) {
      nextProps.load(nextPageUserId, nextIsMyPage);

      return {
        ...prevState,
        loading: true,
        isMyPage: nextIsMyPage,
        currentPageUserId: nextPageUserId,
      };
    }

    return {
      ...prevState,
      loading: nextLoading,
      currentPageUserId: nextPageUserId,
      isMyPage: nextIsMyPage,
    };
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
      <ItemList
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
      <TransactionList
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
    const { user, classes } = this.props;
    const { tabValue, loading, isMyPage } = this.state;

    if (loading) {
      return <LoadingComponent />;
    }

    return (
      <BasePageContainer>
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
            <Typography className={classes.numSellItems} variant="h6">
              出品数 {user.numSellItems}
            </Typography>
          </Grid>
        </Grid>
        <Divider variant="middle" />
        <Tabs value={tabValue} onChange={this._handleChange}>
          <Tab label="出品商品" id="tab--item-list" />
          {isMyPage && <Tab label="取引一覧" id="tab--item-list" />}
        </Tabs>
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
