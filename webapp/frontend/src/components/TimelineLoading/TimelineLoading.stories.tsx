import * as React from 'react';
import { storiesOf } from '@storybook/react';
import { TimelineLoading } from '.';

const stories = storiesOf('components/TimelineLoading', module);

stories.add('default', () => <TimelineLoading />);
