import React from 'react';
import { TransactionItem } from '../../dataObjects/item';
import makeStyles from '@material-ui/core/styles/makeStyles';
import GridList from '@material-ui/core/GridList';
import GridListTile from '@material-ui/core/GridListTile';
import InfiniteScroll from 'react-infinite-scroller';
import TransactionContainer from '../../containers/TransactionContainer';
import { Theme } from '@material-ui/core';
import { TimelineLoading } from '../TimelineLoading';

const useStyles = makeStyles((theme: Theme) => ({
  grid: {
    width: '900px',
    height: '300px',
  },
  tile: {
    overflow: 'inherit',
  },
}));

interface Props {
  items: TransactionItem[];
  hasNext: boolean;
  loadMore: (createdAt: number, itemId: number, page: number) => void;
}

const TransactionList: React.FC<Props> = function({
  items,
  hasNext,
  loadMore,
}: Props) {
  const classes = useStyles();

  const transactionsComponents = [];

  for (const item of items) {
    transactionsComponents.push(
      <GridListTile
        className={classes.grid}
        classes={{ tile: classes.tile }}
        key={item.id}
      >
        <TransactionContainer item={item} />
      </GridListTile>,
    );
  }

  const lastItem = items[items.length - 1];

  return (
    <InfiniteScroll
      pageStart={0}
      loadMore={loadMore.bind(null, lastItem.createdAt, lastItem.id)}
      hasMore={hasNext}
      loader={<TimelineLoading />}
    >
      <GridList cols={1} cellHeight="auto" spacing={6}>
        {transactionsComponents}
      </GridList>
    </InfiniteScroll>
  );
};

export { TransactionList };
