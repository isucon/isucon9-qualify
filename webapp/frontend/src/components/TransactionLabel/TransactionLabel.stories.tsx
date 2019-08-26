import * as React from 'react';
import { storiesOf } from '@storybook/react';
import { TransactionLabel } from '.';

const stories = storiesOf('components/TransactionLabel', module);

stories
  .add('on_sale', () => <TransactionLabel itemStatus={'on_sale'} />)
  .add('trading', () => <TransactionLabel itemStatus={'trading'} />)
  .add('sold_out', () => <TransactionLabel itemStatus={'sold_out'} />)
  .add('stop', () => <TransactionLabel itemStatus={'stop'} />)
  .add('cancel', () => <TransactionLabel itemStatus={'cancel'} />);
