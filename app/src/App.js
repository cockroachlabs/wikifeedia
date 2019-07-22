import React from 'react';
// import logo from './logo.svg';
import './App.css';

import { ApolloClient } from 'apollo-client';
import { HttpLink } from 'apollo-link-http';
import { InMemoryCache } from 'apollo-cache-inmemory';

import gql from 'graphql-tag';
import { Query } from 'react-apollo';

import { ApolloProvider } from 'react-apollo';

const client = new ApolloClient({
  // By default, this client will send queries to the
  //  `/graphql` endpoint on the same host
  // Pass the configuration option { uri: YOUR_GRAPHQL_API_URL } to the `HttpLink` to connect
  // to a different host
  link: new HttpLink({ uri: "/graphqlhttp" }),
  cache: new InMemoryCache(),
});

function Feed() {
  const GET_FEED = gql`
{
  articles {
    abstract
    article
    articleURL
    dailyViews
    imageURL
    thumbnailURL
    title
  }
}`;
  return (<Query query={GET_FEED}>
    {({ loading, error, data }) => {
      if (loading) return 'Loading...';
      if (error) return `Error! ${error.message}`;
      const rows = data.articles.map(({
        article,
        abstract,
        articleURL,
        dailyViews,
        imageURL,
        thumbnailURL,
        title
      }) => (
        <div className="Article" id={article} key={title}>
          <div className="ArticleImageContainer">
            <div className="ArticleImage">
              <a href={articleURL} target="_blank" rel="noopener noreferrer">
                <img src={thumbnailURL} alt={title}/>
              </a>
            </div>
          </div>
          <div className="ArticleContent">
            <h2 className="ArticleTitle">{title}</h2>
            <p className="ArticleAbstractText">
              {abstract}
            </p>
          </div>
        </div>
      ))
      return (
        <div className="Articles">
          {rows}
         </div>
      );
    }}
  </Query>);
}

function App() {
  function App() {
    return (
    <div className="App">
      <div className="AppHeaderContainer">
        <div className="App-header">
            <h1 className="AppTitle">Wikifeedia</h1> 
        </div>
      </div> 
      <Feed/>
    </div>
    );
  };
  return (
    <ApolloProvider client={client}>
      <App />
    </ApolloProvider>
  );
}

export default App;

