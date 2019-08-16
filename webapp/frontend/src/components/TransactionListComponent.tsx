import React from 'react';
import { TransactionItem } from '../dataObjects/item';
import makeStyles from '@material-ui/core/styles/makeStyles';
import GridList from '@material-ui/core/GridList';
import GridListTile from '@material-ui/core/GridListTile';
import InfiniteScroll from 'react-infinite-scroller';
import CircularProgress from '@material-ui/core/CircularProgress/CircularProgress';
import { TransactionComponent } from './TransactionComponent';

const useStyles = makeStyles(theme => ({
  grid: {
    width: '900px',
    height: '300px',
  },
}));

interface Props {
  items: TransactionItem[];
  hasNext: boolean;
  loadMore: (createdAt: number, itemId: number, page: number) => void;
}

const TransactionListComponent: React.FC<Props> = function({
  items,
  hasNext,
  loadMore,
}: Props) {
  const classes = useStyles();

  const transactionsComponents = [];

  for (const item of items) {
    transactionsComponents.push(
      <GridListTile className={classes.grid} key={item.id}>
        <TransactionComponent item={item} />
      </GridListTile>,
    );
  }

  const lastItem = items[items.length - 1];

  return (
    <InfiniteScroll
      pageStart={0}
      loadMore={loadMore.bind(null, lastItem.createdAt, lastItem.id)}
      hasMore={hasNext}
      loader={<CircularProgress />}
    >
      <GridList cols={1}>{transactionsComponents}</GridList>
    </InfiniteScroll>
  );
};

export { TransactionListComponent };
