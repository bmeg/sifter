
import React, { Component } from 'react';
import {Row, Col, ProgressBar} from 'react-materialize'

import './App.css';

class App extends Component {
  render() {
    return (
      <div className="App">
      <Row>
        <Col s={12}>
          <ProgressBar progress={70}/>
        </Col>
      </Row>
      </div>
    );
  }
}

export default App;
